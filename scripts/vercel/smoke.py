# SPDX-License-Identifier: AGPL-3.0-or-later
"""Exercise two built function processes against an isolated, temporary libSQL DB."""
import argparse
import io
import sqlite3
import threading
import zipfile
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import urlsplit, parse_qs
from xml.etree import ElementTree as ET
import json
import os
import secrets
import socket
import subprocess
import tempfile
import time
from contextlib import ExitStack
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path
from urllib.error import HTTPError
from urllib.request import Request, urlopen, build_opener, HTTPRedirectHandler


def port():
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        return sock.getsockname()[1]


def start(stack, args, env, cwd, listen):
    process = subprocess.Popen(args, env=env, cwd=cwd)
    def stop():
        process.terminate()
        process.wait(timeout=10)
    stack.callback(stop)
    for _ in range(150):
        if process.poll() is not None:
            raise RuntimeError("test process exited before listening")
        try:
            with socket.create_connection(("127.0.0.1", listen), timeout=.2):
                return
        except OSError:
            time.sleep(.1)
    raise TimeoutError("test process did not start")


class NoRedirect(HTTPRedirectHandler):
    def redirect_request(self, *args):
        return None


def signed_location(request, status):
    try:
        build_opener(NoRedirect).open(request, timeout=10)
    except HTTPError as redirect:
        with redirect:
            assert redirect.code == status
            assert "no-store" in redirect.headers["Cache-Control"]
            return redirect.headers["Location"]
    raise AssertionError("missing signed S3 download")


def call(listen, path, body=None, expected=200, headers=None, method=None):
    data = None if body is None else json.dumps(body).encode()
    request = Request(f"http://127.0.0.1:{listen}{path}", data=data, method=method,
                      headers={"Content-Type": "application/json", **(headers or {})})
    try:
        response = urlopen(request, timeout=15)
    except HTTPError as error:
        response = error
    with response:
        allowed = (200, 204) if expected is None else expected if isinstance(expected, tuple) else (expected,)
        assert response.status in allowed, (path, response.status, expected)
        return response.read()


def object_fixture(stack):
    # Private local S3 test double. Verify the resulting ZIP and short signed
    # redirect separately; the hosted provider has its own acceptance test.
    objects = {}
    content_types = {}
    uploads = {}
    fail_part_once = {1}
    fail_media_delete_once = {True}

    class Objects(BaseHTTPRequestHandler):
        def log_message(self, *args):
            pass

        def reply(self, body=b"", headers=None):
            self.send_response(200)
            self.send_header("Content-Length", str(len(body)))
            for key, value in (headers or {}).items():
                self.send_header(key, value)
            self.end_headers()
            self.wfile.write(body)

        def do_PUT(self):
            parsed = urlsplit(self.path)
            query = parse_qs(parsed.query)
            body = self.rfile.read(int(self.headers["Content-Length"]))
            if "partNumber" in query:
                number = int(query["partNumber"][0])
                if number in fail_part_once:
                    fail_part_once.remove(number)
                    self.send_error(400, "injected recoverable write failure")
                    return
                uploads[query["uploadId"][0]][number] = body
                return self.reply(headers={"ETag": f'"part-{number}"'})
            if self.headers.get("If-None-Match") == "*" and parsed.path in objects:
                self.send_error(412)
                return
            content_types[parsed.path] = self.headers.get("Content-Type")
            objects[parsed.path] = body
            self.reply()

        def do_HEAD(self):
            path = urlsplit(self.path).path
            if path not in objects or not self.headers.get("Authorization", "").startswith("AWS4-"):
                return self.send_error(404)
            self.send_response(200)
            self.send_header("Content-Length", str(len(objects[path])))
            self.send_header("Content-Type", content_types.get(path) or "application/octet-stream")
            self.end_headers()

        def do_DELETE(self):
            path = urlsplit(self.path).path
            if content_types.get(path) == "image/png" and fail_media_delete_once:
                fail_media_delete_once.clear()
                return self.send_error(400, "injected recoverable object deletion failure")
            objects.pop(path, None)
            self.reply()

        def do_POST(self):
            parsed = urlsplit(self.path)
            if parsed.path == "/v1/embeddings":
                req = json.loads(self.rfile.read(int(self.headers["Content-Length"])))
                assert req["model"] == "fixture" and req["dimensions"] == 2
                embeddings = [{"object": "embedding", "embedding": [3, 4], "index": n}
                              for n, _ in enumerate(req["input"])]
                return self.reply(json.dumps({"object": "list", "data": embeddings}).encode(),
                                  {"Content-Type": "application/json"})
            query = parse_qs(parsed.query, keep_blank_values=True)
            if "delete" in query:
                root = ET.fromstring(self.rfile.read(int(self.headers["Content-Length"])))
                for key in root.findall(".//{*}Key"):
                    objects.pop("/testing/" + key.text, None)
                return self.reply(b"<DeleteResult/>")
            if "uploads" in query:
                upload = secrets.token_hex(8)
                uploads[upload] = {}
                body = f"<InitiateMultipartUploadResult><UploadId>{upload}</UploadId></InitiateMultipartUploadResult>"
            else:
                upload = query["uploadId"][0]
                root = ET.fromstring(self.rfile.read(int(self.headers["Content-Length"])))
                numbers = [int(p.text) for p in root.findall(".//{*}PartNumber")]
                parts = uploads.pop(upload)
                assert all(len(parts[i]) >= 5 << 20 for i in numbers[:-1])
                objects[parsed.path] = b"".join(parts[i] for i in numbers)
                body = "<CompleteMultipartUploadResult><ETag>complete</ETag></CompleteMultipartUploadResult>"
            self.reply(body.encode())

        def do_GET(self):
            path = urlsplit(self.path)
            if "list-type=2" in path.query:
                result = ET.Element("ListBucketResult")
                prefix = parse_qs(path.query).get("prefix", [""])[0]
                for key, data in list(objects.items()):
                    name = key.split("/", 2)[2]
                    if not name.startswith(prefix):
                        continue
                    entry = ET.SubElement(result, "Contents")
                    ET.SubElement(entry, "Key").text = name
                    ET.SubElement(entry, "Size").text = str(len(data))
                body = ET.tostring(result)
            elif path.path in objects and ("X-Amz-Signature=" in path.query or self.headers.get("Authorization", "").startswith("AWS4-")):
                body = objects[path.path]
            else:
                self.send_error(403)
                return
            self.reply(body)

    server = ThreadingHTTPServer(("127.0.0.1", 0), Objects)
    threading.Thread(target=server.serve_forever, daemon=True).start()
    stack.callback(server.server_close)
    stack.callback(server.shutdown)
    return server.server_port


def sql(db_port, statement):
    body = {"requests": [{"type": "execute", "stmt": {"sql": statement}}, {"type": "close"}]}
    result = json.loads(call(db_port, "/v2/pipeline", body))["results"][0]
    assert result["type"] == "ok", result
    return result["response"]["result"]["rows"]


def verify_media(first, second, token, db_port, secret):
    headers = {"Authorization": "Bearer " + token}
    png = bytes.fromhex("89504e470d0a1a0a") + b"fixture"
    data = {"file_name": "image.png", "content_type": "image/png", "category": "image", "size": len(png)}
    media = []
    for _ in range(2):
        presign = json.loads(call(first, "/api/files/presign", data, method="PUT", headers=headers))["data"]
        upload = Request(presign["presign_url"], png, {"Content-Type": "image/png", "If-None-Match": "*"}, method="PUT")
        with urlopen(upload, timeout=5) as response:
            assert response.status == 200
        meta = "/api/file/" + presign["id"] + "/meta"
        for target in (second, first):
            finalized = json.loads(call(target, meta, {"size": len(png)}, method="PUT", headers=headers))
            assert finalized["code"] == 1, finalized
        media.append((presign["id"], "/media/" + presign["key"]))
    orphan, kept = media
    call(first, kept[1], expected=404)
    private_url = signed_location(Request(f"http://127.0.0.1:{second}" + kept[1], headers=headers), 307)
    assert "X-Amz-Expires=120" in private_url
    with urlopen(private_url, timeout=5) as response:
        assert response.read() == png
    result = json.loads(call(first, "/api/echo", {"content": "media fixture", "private": False,
                             "echo_files": [{"file_id": kept[0]}]}, headers=headers))
    assert result["code"] == 1, result
    assert call(second, kept[1]) == png
    sql(db_port, "UPDATE temp_files SET expire_at=1")
    worker = {"Authorization": "Bearer " + secret}
    tick = {"due": int(time.time()) // 3600 * 3600}
    call(first, "/api/maintenance", tick, expected=503, headers=worker)
    assert sql(db_port, "SELECT 1 FROM upload_intents WHERE expires_at=0")
    call(second, "/api/maintenance", tick, expected=None, headers=worker)
    call(first, "/api/maintenance", tick, expected=None, headers=worker)
    call(second, orphan[1], headers=headers, expected=404)
    assert call(second, kept[1]) == png
    assert sql(db_port, "SELECT 1 FROM upload_intents") == []
    call(first, "/api/echo/" + result["data"]["id"], headers=headers, method="DELETE")


def verify_effects(first, second, token, db_port, secret):
    header = {"Authorization": "Bearer " + token}
    worker = {"Authorization": "Bearer " + secret}
    sql(db_port, "CREATE TRIGGER fixture_rollback BEFORE INSERT ON ech0_deliveries BEGIN SELECT RAISE(ABORT,'fixture rollback'); END")
    call(first, "/api/echo", {"content": "rollback probe","private": True}, expected=400, headers=header)
    assert sql(db_port, "SELECT 1 FROM echos WHERE content='rollback probe'") == []
    logs = json.loads(call(second, "/api/system/logs?level=error&keyword=fixture%20rollback", headers=header))
    assert logs["code"] == 1 and any("fixture rollback" in entry["raw"] for entry in logs["data"]), logs
    stream = Request(f"http://127.0.0.1:{second}/api/system/logs/stream?token={token}&level=error&keyword=fixture%20rollback",
                     headers={"Last-Event-ID": "0"})
    with urlopen(stream, timeout=5) as response:
        assert response.readline().startswith(b"id: ")
        entry = json.loads(response.readline().removeprefix(b"data: "))
        assert entry["code"] == 1 and "fixture rollback" in entry["data"]["raw"]
    sql(db_port, "DROP TRIGGER fixture_rollback")
    created = json.loads(call(first, "/api/echo", {"content": "before", "private": True}, headers=header))
    assert created["code"] == 1, created
    echo_id = created["data"]["id"]
    updated = json.loads(call(second, "/api/echo", {"content": "after", "private": True, "id": echo_id}, headers=header, method="PUT"))
    assert updated["data"]["content"] == "after"
    deleted = json.loads(call(first, "/api/echo/" + echo_id, headers=header, method="DELETE"))
    assert deleted["code"] == 1
    reconcile = {"due": int(time.time()) // 3600 * 3600}
    result = json.loads(call(second, "/api/maintenance", reconcile, headers=worker))
    ticket = result[0]
    seen = set()
    while ticket:
        assert ticket["effect"] not in seen
        seen.add(ticket["effect"])
        response = call(first, "/api/maintenance", ticket, headers=worker, expected=None)
        replay = call(second, "/api/maintenance", ticket, headers=worker, expected=None)
        assert replay == response
        ticket = json.loads(response) if response else None
    assert len(seen) == 3 and sql(db_port, "SELECT 1 FROM ech0_deliveries") == []


def verify_reindex(first, second, token, db_port, storage_port, secret):
    auth = {"Authorization": "Bearer " + token}
    worker = {"Authorization": "Bearer " + secret}
    due = int(time.time()) // 3600 * 3600
    config = json.dumps({"model": "fixture", "enable": True, "dim": 2, "api_key": "test",
                         "base_url": f"http://127.0.0.1:{storage_port}/v1"})
    sql(db_port, f"UPDATE key_values SET value='{config}' WHERE key='embedding_setting'")
    sql(db_port, "INSERT INTO echos(id,content,user_id,private) SELECT 'fixture-index','index me',id,1 FROM users")
    def start():
        result = json.loads(call(first, "/api/embedding/reindex", {}, headers=auth))
        assert result["data"]["status"] == "pending", result
        return json.loads(call(second, "/api/maintenance", {"due": due}, headers=worker))[0]
    abandoned = start()
    cancelled = json.loads(call(second, "/api/embedding/reindex/cancel", {}, headers=auth))
    assert cancelled["data"]["status"] == "cancelled"
    assert not call(first, "/api/maintenance", abandoned, headers=worker, expected=None)
    step = start()
    for n in range(8):
        process = first if n % 2 else second
        response = call(process, "/api/maintenance", step, headers=worker, expected=None)
        assert call(first, "/api/maintenance", step, headers=worker, expected=None) == response
        if not response:
            break
        step = json.loads(response)
    state = json.loads(call(second, "/api/embedding/reindex/status", headers=auth))["data"]
    assert state["status"] == "success" and state["payload"]["indexed"] == 1, state
    distance = sql(db_port, "SELECT vector_distance_l2(embedding,vector32('[0,0]')) FROM vec_echo WHERE echo_id='fixture-index'")
    assert float(distance[0][0]["value"]) == 5


def verify_web_export(first, second, token, root, storage_port, db_port, cron_secret, username):
    header = {"Authorization": "Bearer " + token}
    # Exercise multipart and independent checkpoints without contacting any hosted database.
    sql(db_port, "WITH RECURSIVE n(i) AS (VALUES(1) UNION ALL SELECT i+1 FROM n WHERE i<160) INSERT INTO key_values(key,value) SELECT 'export-fixture-'||i,hex(randomblob(32768)) FROM n")
    started = json.loads(call(first, "/api/migration/export", {"format": "snapshot"}, headers=header))
    assert started["code"] == 1, started
    worker = {"Authorization": "Bearer " + cron_secret}
    due = int(time.time()) // 3600 * 3600
    next_step = json.loads(call(second, "/api/maintenance", {"due": due}, headers=worker))
    for index in range(160):
        process = first if index % 2 else second
        try:
            response = call(process, "/api/maintenance", next_step, headers=worker, expected=None)
        except AssertionError as error:
            if not error.args or error.args[0][1] != 503:
                raise
            # Same committed cursor on a different function after a failed part.
            response = call(second if process == first else first, "/api/maintenance", next_step, headers=worker, expected=None)
        if not response:
            break
        updated = json.loads(response)
        if index == 1:
            assert json.loads(call(process, "/api/maintenance", next_step, headers=worker)) == updated
            sql(db_port, "UPDATE key_values SET value='later' WHERE key='export-fixture-1'")
        next_step = updated
    else:
        raise AssertionError("worker did not finish the snapshot")
    for _ in range(120):
        state = json.loads(call(second, "/api/migration/export/status", headers=header))["data"]
        if state["status"] == "success":
            break
        assert state["status"] in ("pending", "running"), state
        time.sleep(.1)
    else:
        raise TimeoutError("snapshot did not finish")

    request = Request(f"http://127.0.0.1:{second}/api/migration/export/download", headers=header)
    location = signed_location(request, 302)
    assert urlsplit(location).port == storage_port and "X-Amz-Expires=300" in location
    link = json.loads(call(second, "/api/migration/export/download?format=snapshot&link=1", headers=header))
    assert link["code"] == 1 and urlsplit(link["data"]).path == urlsplit(location).path
    with urlopen(location, timeout=10) as response, zipfile.ZipFile(io.BytesIO(response.read())) as archive:
        assert archive.getinfo("ech0.db").file_size > 8 << 20
        assert archive.namelist() == ["ech0.db"]
        path = Path(root) / "restored.db"
        path.write_bytes(archive.read("ech0.db"))
    with sqlite3.connect(path) as restored:
        assert restored.execute("select username from users").fetchall() == [(username,)]
        assert restored.execute("select length(value) from key_values where key='export-fixture-1'").fetchone() == (65536,)
        assert restored.execute("pragma integrity_check").fetchone() == ("ok",)


def verify_device_flow(first, second, session):
    headers = {"Authorization": "Bearer " + session}
    started = json.loads(call(first, "/api/auth/device/code", {"client_name": "test CLI", "scopes": ["echo:read"]}))["data"]
    pending = json.loads(call(second, "/api/auth/device/pending?user_code=" + started["user_code"], headers=headers))
    assert pending["data"]["client_name"] == "test CLI"
    data = {"device_code": started["device_code"]}
    call(second, "/api/auth/device/token", data, expected=400)
    approved = json.loads(call(second, "/api/auth/device/approve", {"user_code": started["user_code"]}, headers=headers))
    assert approved["code"] == 1
    grant = json.loads(call(first, "/api/auth/device/token", data))["data"]
    assert grant["scopes"] == ["echo:read"] and grant["access_token"]
    call(second, "/api/auth/device/token", data, expected=400)


def run(args):
    with tempfile.TemporaryDirectory() as root, ExitStack() as stack:
        db_port = port()
        storage_port = object_fixture(stack)
        env = {k: v for k, v in os.environ.items() if not k.startswith(("ECH0_", "TURSO_"))}
        env.update(ECH0_DB_TYPE="libsql", TURSO_DATABASE_URL=f"http://127.0.0.1:{db_port}",
                   JWT_SECRET=secrets.token_hex(32), CRON_SECRET=secrets.token_hex(32),
                   ECH0_LOG_FILE_ENABLE="false", ECH0_OBJECT_ENABLED="true",
                   ECH0_UPLOAD_VERIFIED_DIRECT="true", ECH0_S3_PROVIDER="r2",
                   ECH0_S3_BUCKET="testing", ECH0_S3_ENDPOINT=f"http://127.0.0.1:{storage_port}", ECH0_S3_USE_PATH_STYLE="true",
                   ECH0_S3_REGION="auto", ECH0_S3_ACCESS_KEY="test", ECH0_S3_SECRET_KEY="test",
                   ECH0_STORAGE_DATA_ROOT=str(Path(root) / "objects"))
        start(stack, [args.sqld, "--db-path", "test.db", "--http-listen-addr", f"127.0.0.1:{db_port}"], env, root, db_port)
        subprocess.run([args.cli, "prepare-libsql"], env=env, cwd=root, check=True)
        first, second = port(), port()
        for listen in (first, second):
            start(stack, [args.server], {**env, "PORT": str(listen)}, root, listen)
            call(listen, "/healthz")
            assert b"<html" in call(listen, "/").lower()
            assert "paths" in json.loads(call(listen, "/api/openapi.json"))
            call(listen, "/.well-known/missing", expected=404)
            call(listen, "/api/maintenance", expected=401)
            catalog = json.loads(call(listen, "/api/_catalog"))
            assert catalog["ops"]["file.upload"]["transport"]["request"]["direct_upload"]
        account = {"username": "tester", "password": secrets.token_hex(16), "email": "tester@example.test"}
        other = {**account, "username": "other-owner", "email": "other@example.test"}
        with ThreadPoolExecutor(max_workers=2) as executor:
            tries = [executor.submit(call, listen, "/api/init/owner", candidate, (200, 400, 403, 409))
                     for listen, candidate in ((first, account), (second, other))]
            accepted = [candidate for result, candidate in zip(tries, (account, other))
                        if json.loads(result.result())["code"] == 1]
        assert len(accepted) == 1, "concurrent initialization must produce a single owner"
        account = accepted[0]
        assert json.loads(call(second, "/api/init/status"))["data"]["owner_exists"]
        auth = json.loads(call(second, "/api/login", account))
        assert auth["code"] == 1
        verify_device_flow(first, second, auth["data"]["access_token"])
        verify_effects(first, second, auth["data"]["access_token"], db_port, env["CRON_SECRET"])
        verify_reindex(first, second, auth["data"]["access_token"], db_port, storage_port, env["CRON_SECRET"])
        verify_web_export(first, second, auth["data"]["access_token"], root, storage_port, db_port, env["CRON_SECRET"], account["username"])
        verify_media(first, second, auth["data"]["access_token"], db_port, env["CRON_SECRET"])
        print("OK: concurrent ownership, login, device flow, transactional effects, reindex, logs, private snapshot, direct media and orphan collection across two processes")


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    for name in ("sqld", "server", "cli"):
        parser.add_argument("--" + name, required=True, type=Path)
    run(parser.parse_args())
