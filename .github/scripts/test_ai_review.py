import json
import tempfile
import unittest
from pathlib import Path
from types import SimpleNamespace
from unittest.mock import patch

import ai_review


class FakeResponse:
    def __init__(self, value):
        self.value = json.dumps(value).encode("utf-8")

    def __enter__(self):
        return self

    def __exit__(self, *_):
        return None

    def read(self):
        return self.value


class InferTests(unittest.TestCase):
    def test_disables_thinking_and_saves_structured_content(self):
        response = {"message": {"content": '{"findings": []}'}}
        with tempfile.TemporaryDirectory() as directory:
            prompt = Path(directory) / "prompt.txt"
            output = Path(directory) / "review.json"
            prompt.write_text("review prompt", encoding="utf-8")
            args = SimpleNamespace(
                model="gpt-oss:20b",
                prompt=str(prompt),
                output=str(output),
            )
            with patch("ai_review.urllib.request.urlopen", return_value=FakeResponse(response)) as urlopen:
                ai_review.infer(args)

            request = urlopen.call_args.args[0]
            body = json.loads(request.data)
            self.assertIs(body["think"], False)
            self.assertEqual(output.read_text(encoding="utf-8"), '{"findings": []}')

    def test_empty_content_reports_response_metadata(self):
        response = {
            "done_reason": "stop",
            "eval_count": 12,
            "message": {"content": "", "thinking": "reasoning"},
        }
        with tempfile.TemporaryDirectory() as directory:
            prompt = Path(directory) / "prompt.txt"
            prompt.write_text("review prompt", encoding="utf-8")
            args = SimpleNamespace(
                model="gpt-oss:20b",
                prompt=str(prompt),
                output=str(Path(directory) / "review.json"),
            )
            with patch("ai_review.urllib.request.urlopen", return_value=FakeResponse(response)):
                with self.assertRaisesRegex(ValueError, "thinking_chars=9"):
                    ai_review.infer(args)


if __name__ == "__main__":
    unittest.main()
