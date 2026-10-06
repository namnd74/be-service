# Backend

Service Go cho GitOps demo, làm việc trên `main`.

| Endpoint | Chức năng |
| --- | --- |
| `/` | Giao diện trạng thái |
| `/healthz` | Kiểm tra health |
| `/version` | Version, source commit và môi trường |

## Chạy backend riêng

Cần Go và Bash. Tạo cấu hình local từ mẫu:

```bash
cp .env.local.example .env.local
```

Chỉnh password và các giá trị riêng trong `.env.local`, rồi chạy trong Bash:

```bash
set -a
source .env.local
set +a
go run .
```

Mặc định HTTP tại <http://localhost:8080/>. `.env.local` không được commit
hoặc gửi vào Docker build context.

| Biến | Chức năng |
| --- | --- |
| `PORT` | Port backend |
| `APP_ENV` | Tên môi trường |
| `DB_PASSWORD` | Secret local |
| `DEMO_MODE` | Bật thao tác demo, chỉ có hiệu lực ở dev |
| `DEMO_FAULT` | Mô phỏng lỗi khi demo mode được bật |

## Demo GitOps tự động với Kubernetes local

Clone repository manifest cạnh repo này thành thư mục `gitops-manifests`,
sau đó theo [README manifest](../gitops-manifests/README.md).
Script Bash dựng k3d/Argo CD/Sealed Secrets và cấu hình ba môi trường trên
`main`. Merge PR backend tự chạy CI để phát hành GHCR và tạo PR image trong
repo manifest. Merge PR manifest khiến Argo tự triển khai dev/staging/prod.
Secret Kubernetes được bootstrap riêng, không lấy password từ env chạy Go riêng.

## Kiểm tra

```bash
bash scripts/check-quality.sh
python3 -m unittest discover -s scripts/tests
```

Quality gồm test, coverage, race, vet và gofmt.

## CI và phát hành

CI mặc định test/build/scan. Trivy chặn HIGH/CRITICAL, xuất report và SBOM.
Khi bật `ENABLE_GITOPS_RELEASE=true` trong GitHub Variables, image đã scan được
publish GHCR, ký keyless Cosign, attest provenance và mở PR digest vào `main`
của repo manifest. Các Actions được pin theo commit SHA. CI không có kubeconfig.

Cấu hình repository và credential riêng qua GitHub Variables/Secrets:
`CONFIG_REPO`, `IMAGE_ARCH` và secret `CONFIG_REPO_PAT`.
Xem bảng cấu hình hosted trong [README manifest](../gitops-manifests/README.md).
Mật khẩu/token và đường dẫn máy không được viết vào source hoặc README.
