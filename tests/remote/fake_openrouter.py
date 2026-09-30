import json
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer


class FakeOpenRouter:
    """Answers /chat/completions with `answer` (or `status`), /key with `remaining`."""

    def __init__(self):
        self.bodies = []
        self.status = 200
        self.remaining = 5
        self.answer = {
            "choices": [{"message": {"content": '{"label": "x"}'}}],
            "usage": {"prompt_tokens": 7, "completion_tokens": 3, "cost": 0},
        }
        fake = self

        class Handler(BaseHTTPRequestHandler):
            def do_GET(self):
                payload = {"data": {"free_model_daily_requests": {"remaining": fake.remaining}}}
                self._reply(200, payload)

            def do_POST(self):
                fake.bodies.append(json.loads(self.rfile.read(int(self.headers["Content-Length"]))))
                self._reply(fake.status, fake.answer if fake.status == 200 else {"error": {}})

            def _reply(self, status, payload):
                data = json.dumps(payload).encode()
                self.send_response(status)
                self.send_header("Content-Length", str(len(data)))
                self.end_headers()
                self.wfile.write(data)

            def log_message(self, *args):
                pass

        self.server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        threading.Thread(target=self.server.serve_forever, daemon=True).start()
        self.api = f"http://127.0.0.1:{self.server.server_address[1]}"

    def close(self):
        self.server.shutdown()
