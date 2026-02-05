# GoPlan Deployment Guide

## Prerequisites

- Kubernetes cluster (1.27+)
- kubectl configured with cluster access
- Kustomize (included in kubectl 1.14+)
- Access to container registry (ghcr.io)
- PostgreSQL 16 with pgvector extension
- TLS certificates for ingress

## Environment Setup

### 1. Create Namespace

```bash
kubectl apply -f infrastructure/kubernetes/base/namespace.yaml
```

### 2. Configure Secrets

Create a secrets file based on the template:

```bash
cp infrastructure/kubernetes/base/secrets.yaml secrets-production.yaml
```

Edit `secrets-production.yaml` with actual values:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: goplan-secrets
  namespace: goplan
type: Opaque
stringData:
  database-url: "postgres://user:password@host:5432/goplan?sslmode=require"
  jwt-secret: "your-secure-jwt-secret-min-32-chars"
  openai-api-key: "sk-your-openai-api-key"
```

Apply the secrets:

```bash
kubectl apply -f secrets-production.yaml
```

### 3. Configure TLS Certificates

Using cert-manager (recommended):

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: goplan-production-tls
  namespace: goplan-production
spec:
  secretName: goplan-production-tls
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
    - goplan.yourcompany.com
```

Or manually create the TLS secret:

```bash
kubectl create secret tls goplan-production-tls \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key \
  -n goplan-production
```

## Deployment

### Staging Environment

```bash
# Preview changes
kubectl kustomize infrastructure/kubernetes/overlays/staging

# Apply changes
kubectl kustomize infrastructure/kubernetes/overlays/staging | kubectl apply -f -

# Verify deployment
kubectl get all -n goplan-staging
```

### Production Environment

```bash
# Preview changes
kubectl kustomize infrastructure/kubernetes/overlays/production

# Apply changes
kubectl kustomize infrastructure/kubernetes/overlays/production | kubectl apply -f -

# Verify deployment
kubectl get all -n goplan-production
```

## Database Migration

Run migrations before deploying new API versions:

```bash
# Create a migration job
kubectl run goplan-migrate \
  --image=ghcr.io/your-org/goplan-api:v1.0.0 \
  --restart=Never \
  --env="DATABASE_URL=$DATABASE_URL" \
  -- ./main migrate up

# Check migration status
kubectl logs goplan-migrate

# Clean up
kubectl delete pod goplan-migrate
```

## Monitoring Setup

### Install Prometheus ServiceMonitor

```bash
kubectl apply -f infrastructure/monitoring/servicemonitor.yaml
```

### Install Alert Rules

```bash
kubectl apply -f infrastructure/monitoring/alertrules.yaml
```

### Verify Monitoring

```bash
# Check ServiceMonitor
kubectl get servicemonitor -n goplan

# Check PrometheusRules
kubectl get prometheusrule -n goplan
```

## Health Checks

### API Health

```bash
curl https://goplan.yourcompany.com/health
```

Expected response:
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "database": "connected"
}
```

### Metrics Endpoint

```bash
curl https://goplan.yourcompany.com/metrics
```

## Scaling

### Manual Scaling

```bash
# Scale API
kubectl scale deployment prod-goplan-api --replicas=5 -n goplan-production

# Scale Frontend
kubectl scale deployment prod-goplan-frontend --replicas=5 -n goplan-production
```

### HPA Configuration

The Horizontal Pod Autoscaler is configured to scale based on:
- CPU utilization (target: 70%)
- Memory utilization (target: 80%)

View HPA status:

```bash
kubectl get hpa -n goplan-production
```

## Troubleshooting

### Pod Not Starting

```bash
# Check pod events
kubectl describe pod <pod-name> -n goplan-production

# Check logs
kubectl logs <pod-name> -n goplan-production --previous
```

### Database Connection Issues

```bash
# Verify secret exists
kubectl get secret goplan-secrets -n goplan-production

# Test connection from pod
kubectl exec -it <pod-name> -n goplan-production -- \
  curl -s localhost:8080/health
```

### Ingress Issues

```bash
# Check ingress status
kubectl describe ingress prod-goplan-ingress -n goplan-production

# Check ingress controller logs
kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx
```

## Rollback

### Quick Rollback

```bash
kubectl rollout undo deployment/prod-goplan-api -n goplan-production
```

### Rollback to Specific Version

```bash
# List revisions
kubectl rollout history deployment/prod-goplan-api -n goplan-production

# Rollback to revision
kubectl rollout undo deployment/prod-goplan-api -n goplan-production --to-revision=2
```

## CI/CD Integration

The CI/CD pipeline automatically:

1. Runs tests on every push
2. Builds Docker images on main/master branch
3. Deploys to staging on develop branch
4. Deploys to production on main/master branch

### Required Secrets

Configure these in GitHub repository secrets:

| Secret | Description |
|--------|-------------|
| `KUBE_CONFIG_STAGING` | Base64-encoded kubeconfig for staging |
| `KUBE_CONFIG_PRODUCTION` | Base64-encoded kubeconfig for production |

Generate kubeconfig secret:

```bash
cat ~/.kube/config | base64 -w0
```
