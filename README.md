# Backend cho GitOps seminar

Service Go có giao diện trạng thái, `/healthz`, `/version`.

Luồng code: `feature/* → dev → stg → prod`, qua PR và review.
Merge vào mỗi nhánh phát hành tự chạy CI: test/race/vet → build → Trivy → GHCR → ký/attest →
PR image digest vào **cùng nhánh** repo manifest. Merge PR manifest khiến Argo deploy môi trường tương ứng.
`main` giữ source/CI dùng chung và không tự publish release.

## Chạy riêng trên máy

Cần Go và Bash. Copy `.env.local.example` nếu chưa có env riêng, chỉnh giá trị rồi chạy:

```bash
cp .env.local.example .env.local
set -a
source .env.local
set +a
go run .
```

HTTP mặc định http://localhost:8080/.

| Env | Vai trò |
| --- | --- |
| `PORT` | Port HTTP |
| `APP_ENV` | Tên môi trường |
| `DB_PASSWORD` | Secret; giao diện chỉ báo đã nạp, không hiển thị giá trị |
| `DEMO_MODE` / `DEMO_FAULT` | Mô phỏng health, chỉ có hiệu lực trong dev |

Kubernetes lấy env công khai từ ConfigMap và password từ Secret riêng cho từng môi trường.
Env được truyền lúc chạy; image vẫn được build riêng từng nhánh theo flow seminar.
Không có kết nối database thật.

## CI và cấu hình riêng

GitHub Variables: `ENABLE_GITOPS_RELEASE=true`, `CONFIG_REPO`, `IMAGE_ARCH` (`arm64`/`amd64`).
GitHub Secret: `CONFIG_REPO_PAT` có quyền cập nhật manifest và tạo PR.
Không commit credential hoặc `.env.local`.

Trivy chặn HIGH/CRITICAL (kể cả chưa có bản fix), xuất report/SBOM.
Image publish là image đã scan, deploy bằng digest; Cosign/provenance gắn với nhánh build.
CI không có kubeconfig. Release cũ hơn HEAD của nhánh không được tạo PR cập nhật cấu hình.

`DEMO_FAIL_PROD_BUILD=true` là **GitHub Variable tùy chọn chỉ dùng seminar**:
Docker build trên push/dispatch prod cố ý thất bại; PR checks và dev/stg không bị ảnh hưởng.
Mặc định false. Đặt lại false sau demo. Build failure không đổi image đang chạy.

## Kiểm tra

```bash
bash scripts/check-quality.sh
python3 -m unittest discover -s scripts/tests
```

Clone repo manifest cạnh checkout này. Xem README và `docs/demo.md` của repo manifest
để dựng cluster, chạy release ba môi trường và rollback.
