# Backend dùng cho GitOps demo

Service Go có `/healthz` và `/version`. Repo làm việc trên `main`.
Source không phụ thuộc tài khoản GitHub hoặc đường dẫn máy cụ thể.

Chạy nhanh backend:

```bash
APP_ENV=dev DB_PASSWORD=local-demo DEMO_MODE=true go run .
```

Mở <http://localhost:8080/version>. Password này chỉ là giá trị demo local.
Kiểm tra: `bash scripts/check-quality.sh` (Go cần hỗ trợ các flags trong script).

Để chạy toàn bộ Docker/k3d/Sealed Secrets/Argo CD và ba môi trường local,
clone repo cấu hình cạnh repo này thành thư mục `gitops-manifests`, rồi:

```bash
cd ../gitops-manifests
cp .env.local.example .env.local
bash scripts/local-dev.sh up
```

Xem [hướng dẫn từng bước](../gitops-manifests/scripts/rebuild-step-by-step.md).
Image local build theo kiến trúc Docker host; không cần GHCR hoặc PAT.

CI GitHub mặc định chạy quality, build và scan. Publish/ký/attest và mở PR vào
repo cấu hình chỉ bật khi đặt variable `ENABLE_GITOPS_RELEASE=true`.
`CONFIG_REPO` chỉ định repo manifest, `CONFIG_BRANCH` mặc định `main`,
`IMAGE_ARCH` mặc định `amd64`. Cần secret `CONFIG_REPO_PAT` để tạo PR config.
Xem [runbook cấu hình hosted](../gitops-manifests/scripts/demo-runbook.md) và
[supply-chain controls](docs/supply-chain.md).
