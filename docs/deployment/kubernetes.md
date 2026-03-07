# Kubernetes 部署指南

## 概述

本文档介绍如何在 Kubernetes 集群上部署团圆寻亲系统。

## 前提条件

- Kubernetes 1.24+
- kubectl 已配置
- Helm 3.x (可选)

## 部署步骤

### 1. 创建 Namespace

```bash
kubectl create namespace cntuanyuan
```

### 2. 部署 PostgreSQL

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
  namespace: cntuanyuan
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:15-alpine
        env:
        - name: POSTGRES_USER
          value: postgres
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: password
        - name: POSTGRES_DB
          value: cntuanyuan
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
      volumes:
      - name: postgres-storage
        persistentVolumeClaim:
          claimName: postgres-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: postgres
  namespace: cntuanyuan
spec:
  selector:
    app: postgres
  ports:
  - port: 5432
```

### 3. 部署 Redis

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: cntuanyuan
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: cntuanyuan
spec:
  selector:
    app: redis
  ports:
  - port: 6379
```

### 4. 部署 API 服务

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cntuanyuan-api
  namespace: cntuanyuan
spec:
  replicas: 3
  selector:
    matchLabels:
      app: cntuanyuan-api
  template:
    metadata:
      labels:
        app: cntuanyuan-api
    spec:
      containers:
      - name: api
        image: cntuanyuan-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: postgres
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: password
        - name: REDIS_HOST
          value: redis
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: cntuanyuan-api
  namespace: cntuanyuan
spec:
  selector:
    app: cntuanyuan-api
  ports:
  - port: 80
    targetPort: 8080
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: cntuanyuan-ingress
  namespace: cntuanyuan
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: api.cntuanyuan.org
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: cntuanyuan-api
            port:
              number: 80
```

## 自动扩缩容

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: cntuanyuan-api-hpa
  namespace: cntuanyuan
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: cntuanyuan-api
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## 常用命令

```bash
# 应用配置
kubectl apply -f k8s/

# 查看 Pod 状态
kubectl get pods -n cntuanyuan

# 查看日志
kubectl logs -f deployment/cntuanyuan-api -n cntuanyuan

# 扩缩容
kubectl scale deployment cntuanyuan-api --replicas=5 -n cntuanyuan

# 进入容器
kubectl exec -it <pod-name> -n cntuanyuan -- /bin/sh
```
