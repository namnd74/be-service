# CI và supply chain

Workflow build một image, scan đúng image đó, rồi chỉ sau khi gate đạt mới
publish lên GHCR trên `main`. Image dùng tag bất biến `sha-<full commit SHA>` và
digest registry; không dùng `latest`. `release.json` chỉ được giữ như artifact
của CI cùng report vulnerability, CycloneDX SBOM và metadata run.

Trivy chặn gate khi có vulnerability `HIGH` hoặc `CRITICAL`, dù đã có bản sửa
hay chưa, đồng thời xuất JSON report và SBOM. Trên `main`, sau khi publish,
image nhận GitHub artifact attestation (SLSA provenance) và chữ ký keyless Cosign
trong GHCR. OIDC identity
của GitHub Actions là ngắn hạn, không có private key trong repo.

PR cập nhật Dev do CI tạo sau gate. Promotion `dev -> staging` và
`staging -> prod` dùng cùng digest và tạo PR; workflow từ chối cặp môi trường
không được hỗ trợ, image mutable, digest không khớp hoặc gate/provenance không
hợp lệ. Rollback cũng tạo PR, khôi phục image và `deployment-env-patch.yaml`
theo một revision Git đã biết là tốt; Sealed Secret không bị thay đổi.

Actions được pin vào commit SHA bất biến, giữ tag phát hành trong comment để
review. Secret `CONFIG_REPO_PAT` phải tồn tại trong cả hai repo, có
contents/pull_requests cho config repo và quyền đọc packages/provenance của BE;
workflow không fallback sang `GITHUB_TOKEN`. Quyền `GITHUB_TOKEN` được cấp theo job; job test chỉ cần contents read,
job publish/attest mới cần quyền package và attest. CI không có kubeconfig.

Để kiểm tra image đã publish:

```sh
docker buildx imagetools inspect ghcr.io/OWNER/be-service:sha-FULL_COMMIT_SHA
docker buildx imagetools inspect ghcr.io/OWNER/be-service@sha256:DIGEST
```

Helper dùng chung tại `gitops-manifests/scripts/verify-image.sh` kiểm tra source,
version, revision trong OCI labels; source
SHA phải là commit của `main`, provenance phải chỉ ra đúng repository, SHA và
workflow identity GitHub, rồi mới xác minh chữ ký Cosign:

```sh
cd /Volumes/MacOs/workspaces/git-ops/gitops-manifests
scripts/verify-image.sh ghcr.io/OWNER/be-service@sha256:DIGEST OWNER/be-service
```

`REF` phải là GHCR image reference kèm digest đầy đủ; `OWNER/APP` là repository GitHub. Helper cần
`ci.yaml` trên `main` và issuer
`https://token.actions.githubusercontent.com`. Có thể kiểm tra attestation
riêng bằng `gh attestation verify`.
