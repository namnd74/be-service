"""Check the real Dev update step fails before mutation for stale releases."""
import os
import pathlib
import subprocess
import tempfile
import textwrap
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[2]
SHA = 'a'*40

class DevUpdateTest(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.root = pathlib.Path(self.directory.name)
        (self.root/'apps/be-service/envs/dev').mkdir(parents=True)
        (self.root/'scripts').mkdir()
        (self.root/'bin').mkdir()
        for name, script in {
            'gh': 'printf "%s\\n" "$REMOTE_SHA"',
            'kustomize': 'printf "update %s\\n" "$*" >> "$CALLS"',
        }.items():
            p = self.root/'bin'/name
            p.write_text('#!/bin/sh\n'+script+'\n')
            p.chmod(0o755)
        (self.root/'scripts/validate-manifests.sh').write_text('echo validate >> "$CALLS"\n')
        workflow = (ROOT/'.github/workflows/ci.yaml').read_text()
        step = workflow.split('      - name: Prepare dev image update\n', 1)[1].split('      - name:', 1)[0]
        self.script = textwrap.dedent(step.split('        run: |\n', 1)[1])
        self.env = dict(os.environ, PATH=str(self.root/'bin')+':'+os.environ['PATH'],
                        CALLS=str(self.root/'calls'), SOURCE_SHA=SHA, REMOTE_SHA=SHA,
                        SOURCE_REPO='namnd74/be-service', IMAGE='ghcr.io/namnd74/be-service',
                        DIGEST='sha256:'+'b'*64)

    def prepare(self):
        return subprocess.run(['bash', '-c', self.script], cwd=self.root, env=self.env,
                              capture_output=True, text=True)

    def test_stale_release_does_not_touch_configuration(self):
        self.env['REMOTE_SHA'] = 'c'*40
        self.assertNotEqual(self.prepare().returncode, 0)
        self.assertFalse((self.root/'calls').exists())

    def test_invalid_digest_does_not_touch_configuration(self):
        self.env['DIGEST'] = 'latest'
        self.assertNotEqual(self.prepare().returncode, 0)
        self.assertFalse((self.root/'calls').exists())

    def test_current_release_updates_only_immutable_reference_and_validates(self):
        result = self.prepare()
        self.assertEqual(result.returncode, 0, result.stderr)
        calls = (self.root/'calls').read_text()
        self.assertIn('edit set image '+self.env['IMAGE']+'='+self.env['IMAGE']+'@'+self.env['DIGEST'], calls)
        self.assertIn('validate', calls)


    def test_invalid_version_is_rejected_before_metadata_output(self):
        workflow = (ROOT/'.github/workflows/ci.yaml').read_text()
        step = workflow.split('      - id: meta\n', 1)[1].split('      - name:', 1)[0]
        script = textwrap.dedent(step.split('        run: |\n', 1)[1])
        (self.root/'VERSION').write_text('v1.2.3$(touch injected)')
        output = self.root/'output'
        result = subprocess.run(['bash', '-c', script], cwd=self.root,
                                env=dict(self.env, GITHUB_OUTPUT=str(output), GITHUB_SHA=SHA),
                                capture_output=True, text=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertFalse((self.root/'injected').exists())
        self.assertFalse(output.exists())

    def test_valid_version_records_metadata(self):
        workflow = (ROOT/'.github/workflows/ci.yaml').read_text()
        step = workflow.split('      - id: meta\n', 1)[1].split('      - name:', 1)[0]
        script = textwrap.dedent(step.split('        run: |\n', 1)[1])
        (self.root/'VERSION').write_text('v1.2.3\n')
        output = self.root/'output'
        result = subprocess.run(['bash', '-c', script], cwd=self.root,
                                env=dict(self.env, GITHUB_OUTPUT=str(output), GITHUB_SHA=SHA),
                                capture_output=True, text=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(output.read_text(), 'version=v1.2.3\nsource_sha='+SHA+'\n')

if __name__ == '__main__':
    unittest.main()
