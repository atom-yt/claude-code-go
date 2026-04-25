# Linux OOM Killer 防护方案

## 问题背景

系统发生频繁 OOM (Out Of Memory)，导致关键系统服务被杀：

| 进程 | 被杀原因 | 内存占用 |
|------|----------|----------|
| node_agent | oom_score_adj=0 | 33 MB |
| dcosAgent | oom_score_adj=0 | 7.7 MB |
| dcosLifecycleMa | oom_score_adj=0 | 7.5 MB |
| dcosMonitorPlug | oom_score_adj=0 | 5 MB |
| supervisor.dcos | oom_score_adj=0 | 2.7 MB |

**根本原因**：`flux_engine.cud` (CUDA 计算进程) 占用大量内存，其 `oom_score_adj=-997` 受保护，内核只好杀掉无保护的系统进程。

---

## 一、进程级别保护

### 方案 1：调整 oom_score_adj（最直接）

```bash
# 查看当前进程的 oom 分数
cat /proc/$(pidof node_agent)/oom_score_adj

# 临时设置：-1000 表示永不杀
echo -1000 > /proc/$(pidof node_agent)/oom_score_adj
echo -1000 > /proc/$(pidof dcosAgent)/oom_score_adj
echo -1000 > /proc/$(pidof supervisor.dcos)/oom_score_adj
```

### 方案 2：创建保护脚本

```bash
# 创建 /usr/local/bin/oom-protect.sh
cat > /usr/local/bin/oom-protect.sh << 'EOF'
#!/bin/bash
# 保护关键进程不被 OOM kill
PROCESSES="node_agent dcosAgent supervisor.dcos dcosLifecycleMa dcosMonitorPlug"

for proc in $PROCESSES; do
    pids=$(pgrep -f "^.*$proc")
    for pid in $pids; do
        if [ -d /proc/$pid ]; then
            echo -1000 > /proc/$pid/oom_score_adj 2>/dev/null
            echo "Protected $proc (PID: $pid)"
        fi
    done
done
EOF

chmod +x /usr/local/bin/oom-protect.sh

# 添加到 crontab，每分钟检查一次
echo "* * * * * /usr/local/bin/oom-protect.sh >> /var/log/oom-protect.log 2>&1" | crontab -
```

---

## 二、Systemd 服务保护

### 方案 3：Systemd OOMScoreAdjust

```bash
# 编辑服务配置
systemctl edit dcos-agent.service

# 添加：
[Service]
OOMScoreAdjust=-1000
```

---

## 三、Cgroup 内存保护

### 方案 4：Cgroup 内存预留（cgroup v1/v2）

```bash
# 创建受保护的 cgroup
mkdir -p /sys/fs/cgroup/memory/protected

# 设置保护值（至少保留的内存）
echo 1G > /sys/fs/cgroup/memory/protected/memory.min
echo 2G > /sys/fs/cgroup/memory/protected/memory.low

# 把进程移到受保护的 cgroup
echo $(pidof node_agent) > /sys/fs/cgroup/memory/protected/cgroup.procs
echo $(pidof dcosAgent) > /sys/fs/cgroup/memory/protected/cgroup.procs
```

---

## 四、内核参数调整

### 方案 5：让分配者承担 OOM 后果

```bash
# 让触发 OOM 的进程自己被 kill，而不是杀其他进程
echo 1 > /proc/sys/vm/oom_kill_allocating_task

# 永久生效
echo "vm.oom_kill_allocating_task=1" >> /etc/sysctl.conf
```

### 方案 6：调整 Overcommit（防止过度分配）

```bash
# 检查当前设置
cat /proc/sys/vm/overcommit_memory
# 0 = 启发式（默认）
# 1 = 总是允许
# 2 = 严格模式

# 设置为严格模式，禁止过度分配内存
echo 2 > /proc/sys/vm/overcommit_memory
echo 80 > /proc/sys/vm/overcommit_ratio  # 最多允许分配 RAM+swap 的 80%

# 永久生效
cat >> /etc/sysctl.conf << EOF
vm.overcommit_memory=2
vm.overcommit_ratio=80
EOF
```

---

## 五、Swap 缓冲

### 方案 7：增加 Swap

```bash
# 检查当前 swap
free -h

# 如果没有或很少，创建 swap 文件
dd if=/dev/zero of=/swapfile bs=1G count=8
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile

# 永久生效
echo "/swapfile swap swap defaults 0 0" >> /etc/fstab

# 调整 swappiness（0 = 不轻易用 swap，100 = 积极 swap）
sysctl vm.swappiness=10
echo "vm.swappiness=10" >> /etc/sysctl.conf
```

---

## 六、进程监控+自动重启（兜底）

### 方案 8：守护进程自动恢复

```bash
# 创建监控脚本
cat > /usr/local/bin/process-guard.sh << 'EOF'
#!/bin/bash
PROCESSES="node_agent:dcos-agent dcosAgent:dcos-agent supervisor.dcos:dcos-agent"

while true; do
    while IFS=: read -r proc_name service; do
        if ! pgrep -f "$proc_name" > /dev/null; then
            echo "[$(date)] $proc_name not running, restarting $service..."
            systemctl restart "$service" 2>/dev/null || \n            /etc/init.d/"$service" restart 2>/dev/null
            sleep 5
        fi
    done <<< "$PROCESSES"
    sleep 10
done
EOF

chmod +x /usr/local/bin/process-guard.sh

# 用 systemd 运行守护进程
cat > /etc/systemd/system/process-guard.service << 'EOF'
[Unit]
Description=Process Guard for Critical Services
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/process-guard.sh
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable process-guard
systemctl start process-guard
```

---

## 七、Kubelet 侧方案

### 方案 9：增加 Kubelet 系统预留资源

编辑 `/home/work/kube_let/conf/kubelet.json`：

```json
{
  "systemReserved": {
    "cpu": "1",
    "memory": "4Gi"    // 从 2Gi 增加到 4Gi
  },
  "kubeReserved": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

然后重启：
```bash
systemctl restart dcos-agent
```

### 方案 10：提高 Kubelet 驱逐阈值

当前配置：
```
--eviction-hard=imagefs.available<10%,memory.available<500Mi,nodefs.available<10%,nodefs.inodesFree<10%
```

修改为：
```json
{
  "evictionHard": {
    "imagefs.available": "10%",
    "memory.available": "2Gi",    // 从 500Mi 提高到 2Gi
    "nodefs.available": "10%",
    "nodefs.inodesFree": "10%"
  }
}
```

### 方案 11：给 flux_engine 所在 Pod 设置内存限制

```bash
# 找到 flux_engine 所在的 Pod
kubectl get pods --all-namespaces --field-selector spec.hostname=$(hostname) -o wide

# 修改 Pod 的资源限制
kubectl patch pod <pod-name> -n <namespace> -p '{
  "spec": {
    "containers": [{
      "name": "<container-name>",
      "resources": {
        "limits": {"memory": "16Gi"}
      }
    }]
  }
}'
```

### 方案 12：设置 Pod QoS 为 Guaranteed

```yaml
# 确保 request == limit
resources:
  requests:
    memory: "16Gi"
    nvidia.com/gpu: "1"
  limits:
    memory: "16Gi"    # 必须 == request
    nvidia.com/gpu: "1"
```

---

## 八、一键部署脚本

### 方案 13：综合保护方案

```bash
# 创建一键部署脚本
cat > /usr/local/bin/setup-oom-protection.sh << 'EOF'
#!/bin/bash

echo "=== 设置 OOM 保护 ==="

# 1. 内核参数：让分配者承担 OOM
echo 1 > /proc/sys/vm/oom_kill_allocating_task
grep -q "vm.oom_kill_allocating_task" /etc/sysctl.conf || \n    echo "vm.oom_kill_allocating_task=1" >> /etc/sysctl.conf

# 2. 保护关键进程
PROCESSES="node_agent dcosAgent supervisor.dcos dcosLifecycleMa dcosMonitorPlug"
for proc in $PROCESSES; do
    pgrep -f "$proc" | xargs -I {} echo -1000 > /proc/{}/oom_score_adj 2>/dev/null
    echo "Protected $proc"
done

# 3. 如果是 systemd 服务，设置 OOMScoreAdjust
for svc in dcos-agent matrix-bios; do
    systemctl set-property $svc OOMScoreAdjust=-1000 2>/dev/null
    echo "Set OOMScoreAdjust=-1000 for $svc"
done

# 4. 设置定期保护
grep -q "oom-protect" /var/spool/cron/root 2>/dev/null || \n    echo "* * * * * /usr/local/bin/oom-protect.sh" >> /var/spool/cron/root

echo "=== 完成 ==="
EOF

chmod +x /usr/local/bin/setup-oom-protection.sh

# 执行
/usr/local/bin/setup-oom-protection.sh
```

---

## 方案对比总结

| 方案 | 效果 | 推荐度 | 适用场景 | 持久性 |
|------|------|--------|----------|--------|
| oom_score_adj=-1000 | 直接保护 | ⭐⭐⭐⭐⭐ | 所有场景 | 需要脚本 |
| cgroup memory.min | 内存预留 | ⭐⭐⭐⭐ | 新内核/cgroup v2 | 永久 |
| vm.oom_kill_allocating_task=1 | 责任归位 | ⭐⭐⭐⭐ | 根本解决 | 永久 |
| 进程监控+重启 | 兜底方案 | ⭐⭐⭐⭐ | 被杀后恢复 | 永久 |
| vm.overcommit=2 | 预防为主 | ⭐⭐⭐ | 严格控制 | 永久 |
| 增加 swap | 延缓触发 | ⭐⭐ | 临时缓解 | 永久 |
| Kubelet systemReserved | 预留资源 | ⭐⭐⭐⭐⭐ | K8s 环境 | 永久 |
| Kubelet eviction 阈值 | 提前驱逐 | ⭐⭐⭐⭐⭐ | K8s 环境 | 永久 |
| Pod memory limit | 根本解决 | ⭐⭐⭐⭐⭐ | 限制应用 | 永久 |
| Guaranteed QoS | 优先级控制 | ⭐⭐⭐⭐ | GPU 任务 | 永久 |

---

## 推荐实施顺序

### 立即执行（止血）
1. 临时设置 `oom_score_adj=-1000` 保护关键进程
2. 运行一键部署脚本 `setup-oom-protection.sh`

### 短期实施（1-2天）
1. 修改 Kubelet 配置，增加 `system-reserved.memory` 到 4Gi
2. 提高 Kubelet eviction 阈值到 2Gi
3. 启用进程守护 `process-guard.service`

### 长期实施（1周内）
1. 给 flux_engine 所在 Pod 设置 memory limit
2. 设置 Pod QoS 为 Guaranteed
3. 配置 cgroup memory.min/memor.low 保护

---

## 验证配置

```bash
# 验证 oom_score_adj 设置
cat /proc/$(pidof node_agent)/oom_score_adj

# 验证内核参数
sysctl vm.oom_kill_allocating_task
sysctl vm.overcommit_memory

# 验证 Kubelet 配置
systemctl status dcos-agent
journalctl -u kubelet -n 50 | grep -i reserved

# 验证 cgroup 设置
cat /sys/fs/cgroup/system.slice/memory.min
cat /sys/fs/cgroup/system.slice/memory.low

# 验证 crontab
crontab -l
```

---

## 注意事项

1. **`oom_score_adj=-1000` 不是 100% 保证**：极端情况下内核仍可能忽略
2. **根本解决还是要限制内存使用大户**：找到并限制 `flux_engine.cud` 的内存占用
3. **重启服务后需要重新设置**：除非使用 systemd OOMScoreAdjust 或守护脚本
4. **增加 swap 要慎重**：swap 会影响性能，只是延缓 OOM 触发
5. **vm.overcommit=2 可能影响某些应用**：严格模式可能拒绝某些正常的内存分配

---

## 参考资料

- [Linux Kernel Documentation: OOM Killer](https://www.kernel.org/doc/html/latest/admin-guide/mm/oom.html)
- [Kubernetes: System Reserved](https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/)
- [systemd.resource-control](https://www.freedesktop.org/software/systemd/man/systemd.resource-control.html)