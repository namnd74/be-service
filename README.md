# Backend Microservice

Repo ứng dụng Go và CI. `main.go` phục vụ UI, `/version`, `/healthz`; VERSION,
commit SHA và các label OCI được ghi vào image lúc build. Secret chỉ hiển thị
`loaded`/`missing`; fault demo kéo dài chỉ bật ở Dev.

CI chạy test/coverage/race/vet/format, build một image theo `IMAGE_ARCH`, Trivy
scan image đó, rồi trên `main` publish đúng image đã scan lên GHCR theo tag SHA
và digest. Sau khi publish, cùng digest được attested và ký bằng Cosign.
Sau gate, CI tạo PR cập nhật cấu hình Dev; CI không push thẳng vào config repo,
không triển khai Kubernetes. Argo CD đọc config repo để deploy sau khi PR được
review và merge.

`release.json` chỉ là artifact của CI, lưu cùng report scan/SBOM/provenance; nó
không được đặt trong overlay và không phải nguồn trạng thái triển khai.

```bash
bash scripts/check-quality.sh
python3 -m unittest discover -s scripts/tests -v
```

PR chạy quality/build/scan. Push `main` chạy pipeline phát hành và tạo Dev PR.
`CONFIG_REPO_PAT` phải được đặt trong Secrets của cả hai repo, có quyền
contents/pull_requests trên config repo và đọc packages/provenance của BE; không
có fallback sang `GITHUB_TOKEN`. `IMAGE_ARCH` mặc định `arm64` cho lab Apple
Silicon; cluster x86 dùng `amd64`. Promotion và rollback đều là PR trong config
repo, dùng cùng digest đã scan.

Các tool local gồm Go, Python, Docker và bộ kiểm tra manifest theo runbook;
CI cài phiên bản đã pin cùng checksum qua `install-ci-tools.sh`.

Đọc [runbook seminar](../gitops-manifests/scripts/demo-runbook.md) để setup,
connect, promotion, demo security/secret/drift và rollback. [Supply-chain
controls](docs/supply-chain.md) mô tả scan, SBOM, provenance, signing và helper
kiểm tra image.
