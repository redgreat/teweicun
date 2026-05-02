先清构建缓存（最有效）
docker builder prune -af
再清理未使用镜像/网络
docker system prune -af
如果还报错：在 Docker Desktop 里把磁盘上限调大
Settings → Resources → Disk image size（调到更大，比如 128GB）
重新构建
docker compose build --no-cache backend