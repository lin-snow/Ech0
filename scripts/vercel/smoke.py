# SPDX-License-Identifier: AGPL-3.0-or-later
"""Exercise two built function processes against an isolated, temporary libSQL DB."""
import argparse
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
from urllib.request import Request, urlopen


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


def run(args):
    with tempfile.TemporaryDirectory() as root, ExitStack() as stack:
        db_port = port()
        env = {k: v for k, v in os.environ.items() if not k.startswith(("ECH0_", "TURSO_"))}
        env.update(ECH0_DB_TYPE="libsql", TURSO_DATABASE_URL=f"http://127.0.0.1:{db_port}",
                   JWT_SECRET=secrets.token_hex(32), CRON_SECRET=secrets.token_hex(32),
                   ECH0_LOG_FILE_ENABLE="false", ECH0_OBJECT_ENABLED="true",
                   ECH0_UPLOAD_VERIFIED_DIRECT="true", ECH0_S3_PROVIDER="r2",
                   ECH0_S3_BUCKET="testing", ECH0_S3_ENDPOINT="127.0.0.1:1",
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
        print("OK: SPA, discovery, catalog, access control, cross-process owner and login")


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    for name in ("sqld", "server", "cli"):
        parser.add_argument("--" + name, required=True, type=Path)
    run(parser.parse_args())
