# Backend Microservice (`be-service`)

Kho mã nguồn ứng dụng Backend dành cho Lập trình viên trong bài Lab & Seminar GitOps.

## 1. Kiến trúc
- Ngôn ngữ: Go 1.21
- Web UI & REST API: `/`, `/version`, `/healthz`, `/simulate-crash`
- Container: Dockerfile multi-stage, non-root user

## 2. Quy trình CI (GitHub Actions)
1. **Quality Gate:** Test & Lint (`go test`, `go vet`, `gofmt`)
2. **Security Gate:** Quét lỗ hổng Docker image bằng **Trivy** (chặn nếu có CVE `CRITICAL`)
3. **Artifact:** Đóng gói và đẩy Docker image lên GitHub Container Registry (`ghcr.io`)
4. **GitOps Trigger:** Tự động commit cập nhật tag mới vào repository cấu hình `gitops-manifests`.
