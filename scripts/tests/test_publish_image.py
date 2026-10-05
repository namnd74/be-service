#!/usr/bin/env python3
import json
import os
import pathlib
import subprocess
import tempfile
import unittest


ROOT = pathlib.Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "scripts" / "publish-image.sh"
SHA = "a" * 40
DIGEST = "sha256:" + "b" * 64


class PublishImageTests(unittest.TestCase):
    def run_script(self, *args, docker_output=DIGEST):
        with tempfile.TemporaryDirectory() as tmp:
            tmp_path = pathlib.Path(tmp)
            pushed = tmp_path / "pushed"
            docker = tmp_path / "docker"
            docker.write_text(
                "#!/bin/sh\n"
                "if [ \"$1\" = push ]; then touch \"$PUSH_MARKER\"; exit 0; fi\n"
                "if [ \"$1\" = buildx ]; then printf '%s\\n' \"$DOCKER_DIGEST\"; exit 0; fi\n"
                "exit 99\n"
            )
            docker.chmod(0o755)
            release = tmp_path / "release.json"
            env = os.environ.copy()
            env.update(
                PATH=f"{tmp}:{env['PATH']}",
                PUSH_MARKER=str(pushed),
                DOCKER_DIGEST=docker_output,
            )
            result = subprocess.run(
                [str(SCRIPT), *args, str(release)],
                cwd=ROOT,
                env=env,
                text=True,
                capture_output=True,
            )
            return result, pushed.exists(), release.read_text() if release.exists() else ""

    def test_bad_metadata_fails_before_push(self):
        cases = [
            (" ", f"sha-{SHA}", SHA, "v1.2.3", "https://example.test/run"),
            (f"sha-{SHA}", f"sha-{SHA}", SHA, "1.2.3", "https://example.test/run"),
            (f"sha-{SHA}", f"sha-{SHA}", SHA, "v1.2.3", "not-a-url"),
        ]
        for image, tag, source_sha, version, url in cases:
            with self.subTest(image=image, version=version, url=url):
                result, pushed, _ = self.run_script(image, tag, source_sha, version, url)
                self.assertNotEqual(result.returncode, 0)
                self.assertFalse(pushed, result.stderr)

    def test_invalid_remote_digest_fails_after_push(self):
        result, pushed, _ = self.run_script(
            "ghcr.io/example/be-service",
            f"sha-{SHA}",
            SHA,
            "v1.2.3",
            "https://example.test/run",
            docker_output="not-a-digest",
        )
        self.assertNotEqual(result.returncode, 0)
        self.assertTrue(pushed)

    def test_success_writes_strict_release_schema(self):
        result, pushed, release_text = self.run_script(
            "ghcr.io/example/be-service",
            f"sha-{SHA}",
            SHA,
            "v1.2.3",
            "https://example.test/run",
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertTrue(pushed)
        data = json.loads(release_text)
        self.assertEqual(
            set(data), {"image", "digest", "source_sha", "version", "source_run_url"}
        )
        self.assertEqual(data["digest"], DIGEST)


if __name__ == "__main__":
    unittest.main()
