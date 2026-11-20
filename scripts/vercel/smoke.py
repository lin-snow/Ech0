# SPDX-License-Identifier: AGPL-3.0-or-later
"""Exercise two built function processes against an isolated, temporary libSQL DB."""
import argparse
import io
import sqlite3
import threading
import zipfile
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import urlsplit
from xml.etree import ElementTree as ET
import json
import os
import secrets
import socket
import subprocess
import tempfile
import time
from contextlib import ExitStack
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


def call(listen, path, body=None, expected=200, headers=None):
    data = None if body is None else json.dumps(body).encode()
    request = Request(f"http://127.0.0.1:{listen}{path}", data=data,
                      headers={"Content-Type": "application/json", **(headers or {})})
    try:
        response = urlopen(request, timeout=15)
    except HTTPError as error:
        response = error
    with response:
        assert response.status == expected, (path, response.status, expected)
        return response.read()


def object_fixture(stack):
    # Private local S3 test double. Verify the resulting ZIP and short signed
    # redirect separately; the hosted provider has its own acceptance test.
    objects = {}

    class Objects(BaseHTTPRequestHandler):
        def log_message(self, *args):
            pass

        def do_PUT(self):
            key = urlsplit(self.path).path
            objects[key] = self.rfile.read(int(self.headers["Content-Length"]))
            self.send_response(200)
            self.end_headers()

        def do_GET(self):
            path = urlsplit(self.path)
            if "list-type=2" in path.query:
                result = ET.Element("ListBucketResult")
                for key, data in objects.items():
                    entry = ET.SubElement(result, "Contents")
                    ET.SubElement(entry, "Key").text = key.split("/", 2)[2]
                    ET.SubElement(entry, "Size").text = str(len(data))
                body = ET.tostring(result)
            elif "X-Amz-Signature=" in path.query and path.path in objects:
                body = objects[path.path]
            else:
                self.send_error(403)
                return
            self.send_response(200)
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)

    server = ThreadingHTTPServer(("127.0.0.1", 0), Objects)
    threading.Thread(target=server.serve_forever, daemon=True).start()
    stack.callback(server.server_close)
    stack.callback(server.shutdown)
    return server.server_port


def verify_web_export(first, second, token, root, storage_port):
    header = {"Authorization": "Bearer " + token}
    started = json.loads(call(first, "/api/migration/export", {"format": "snapshot"}, headers=header))
    assert started["code"] == 1, started
    for _ in range(120):
        state = json.loads(call(second, "/api/migration/export/status", headers=header))["data"]
        if state["status"] == "success":
            break
        assert state["status"] in ("pending", "running"), state
        time.sleep(.1)
    else:
        raise TimeoutError("snapshot did not finish")

    class NoRedirect(HTTPRedirectHandler):
        def redirect_request(self, *args):
            return None

    request = Request(f"http://127.0.0.1:{second}/api/migration/export/download", headers=header)
    try:
        build_opener(NoRedirect).open(request, timeout=10)
    except HTTPError as redirect:
        with redirect:
            assert redirect.code == 302
            location = redirect.headers["Location"]
            assert urlsplit(location).port == storage_port and "X-Amz-Expires=300" in location
    else:
        raise AssertionError("missing signed S3 download")
    with urlopen(location, timeout=10) as response, zipfile.ZipFile(io.BytesIO(response.read())) as archive:
        assert archive.namelist() == ["ech0.db"]
        path = Path(root) / "restored.db"
        path.write_bytes(archive.read("ech0.db"))
    with sqlite3.connect(path) as restored:
        assert restored.execute("select username from users").fetchall() == [("tester",)]
        assert restored.execute("pragma integrity_check").fetchone() == ("ok",)


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
        assert json.loads(call(first, "/api/init/owner", account))["code"] == 1
        assert json.loads(call(second, "/api/init/status"))["data"]["owner_exists"]
        auth = json.loads(call(second, "/api/login", account))
        assert auth["code"] == 1
        verify_web_export(first, second, auth["data"]["access_token"], root, storage_port)
        print("OK: SPA, discovery, login and private web snapshot across two processes")


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    for name in ("sqld", "server", "cli"):
        parser.add_argument("--" + name, required=True, type=Path)
    run(parser.parse_args())
