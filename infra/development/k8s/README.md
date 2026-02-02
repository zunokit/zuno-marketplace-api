# Kubernetes Deployment for New Services

When adding a new service, create a `<service-name>-deployment.yaml` in this directory.

## Template

Replace `<service-name>` and `<port>` with your service details.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: <service-name>
  namespace: dev
spec:
  ports:
    - port: <port>
      targetPort: <port>
  selector:
    app: <service-name>
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: <service-name>
  namespace: dev
spec:
  replicas: 1
  selector:
    matchLabels:
      app: <service-name>
  template:
    metadata:
      labels:
        app: <service-name>
    spec:
      containers:
        - name: <service-name>
          image: <service-name>:latest
          imagePullPolicy: Never
          ports:
            - containerPort: <port>
          env:
            - name: SERVICE_PORT
              value: ":<port>"
            - name: POSTGRES_HOST
              valueFrom:
                configMapKeyRef:
                  name: app-config
                  key: POSTGRES_HOST
            - name: POSTGRES_PORT
              valueFrom:
                configMapKeyRef:
                  name: app-config
                  key: POSTGRES_PORT
            - name: POSTGRES_USER
              valueFrom:
                configMapKeyRef:
                  name: app-config
                  key: POSTGRES_USER
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: app-secrets
                  key: POSTGRES_PASSWORD
            - name: POSTGRES_DATABASE
              valueFrom:
                configMapKeyRef:
                  name: app-config
                  key: POSTGRES_DATABASE
```
