#!/usr/bin/env bash
# ===========================================
# GoPlan Deployment Script
# Manual deployment to Kubernetes cluster
# ===========================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Default values
ENVIRONMENT="${ENVIRONMENT:-staging}"
VERSION="${VERSION:-latest}"
NAMESPACE=""
DRY_RUN="${DRY_RUN:-false}"
SKIP_BUILD="${SKIP_BUILD:-false}"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Print usage
usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Deploy GoPlan to Kubernetes cluster.

OPTIONS:
    -e, --environment   Environment to deploy to (staging|production) [default: staging]
    -v, --version       Version/tag to deploy [default: latest]
    -n, --namespace     Kubernetes namespace [default: goplan-{environment}]
    -d, --dry-run       Perform a dry run without making changes
    -s, --skip-build    Skip building and pushing Docker images
    -h, --help          Show this help message

EXAMPLES:
    $(basename "$0") -e staging -v v1.0.0
    $(basename "$0") -e production -v sha-abc1234
    $(basename "$0") -e staging --dry-run

ENVIRONMENT VARIABLES:
    KUBECONFIG          Path to kubeconfig file
    REGISTRY            Container registry [default: ghcr.io]
    REGISTRY_USER       Registry username
    REGISTRY_PASSWORD   Registry password

EOF
    exit 0
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            -d|--dry-run)
                DRY_RUN="true"
                shift
                ;;
            -s|--skip-build)
                SKIP_BUILD="true"
                shift
                ;;
            -h|--help)
                usage
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                ;;
        esac
    done

    # Validate environment
    if [[ "$ENVIRONMENT" != "staging" && "$ENVIRONMENT" != "production" ]]; then
        log_error "Invalid environment: $ENVIRONMENT. Must be 'staging' or 'production'."
        exit 1
    fi

    # Set default namespace if not provided
    if [[ -z "$NAMESPACE" ]]; then
        if [[ "$ENVIRONMENT" == "production" ]]; then
            NAMESPACE="goplan-prod"
        else
            NAMESPACE="goplan-staging"
        fi
    fi
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install it first."
        exit 1
    fi

    # Check helm
    if ! command -v helm &> /dev/null; then
        log_error "helm is not installed. Please install it first."
        exit 1
    fi

    # Check docker (if building)
    if [[ "$SKIP_BUILD" != "true" ]]; then
        if ! command -v docker &> /dev/null; then
            log_error "docker is not installed. Please install it or use --skip-build."
            exit 1
        fi
    fi

    # Check kubeconfig
    if [[ ! -f "${KUBECONFIG:-$HOME/.kube/config}" ]]; then
        log_error "kubeconfig not found. Please configure kubectl."
        exit 1
    fi

    # Test cluster connection
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
        exit 1
    fi

    log_success "All prerequisites met."
}

# Build and push Docker images
build_and_push() {
    if [[ "$SKIP_BUILD" == "true" ]]; then
        log_info "Skipping Docker build (--skip-build specified)"
        return
    fi

    local registry="${REGISTRY:-ghcr.io}"
    local backend_image="$registry/goplan/goplan-backend:$VERSION"
    local frontend_image="$registry/goplan/goplan-frontend:$VERSION"

    log_info "Building Docker images..."

    # Login to registry if credentials provided
    if [[ -n "${REGISTRY_USER:-}" && -n "${REGISTRY_PASSWORD:-}" ]]; then
        echo "$REGISTRY_PASSWORD" | docker login "$registry" -u "$REGISTRY_USER" --password-stdin
    fi

    # Build backend
    log_info "Building backend image: $backend_image"
    docker build \
        --target production \
        --build-arg VERSION="$VERSION" \
        --build-arg BUILD_TIME="$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
        --build-arg GIT_COMMIT="$(git rev-parse HEAD 2>/dev/null || echo 'unknown')" \
        -t "$backend_image" \
        -f "$PROJECT_ROOT/Dockerfile" \
        "$PROJECT_ROOT"

    # Build frontend
    log_info "Building frontend image: $frontend_image"

    # Set frontend build args based on environment
    local api_url ws_url app_url
    if [[ "$ENVIRONMENT" == "production" ]]; then
        api_url="https://api.goplan.io/api/v1"
        ws_url="wss://api.goplan.io/ws"
        app_url="https://goplan.io"
    else
        api_url="https://api.staging.goplan.io/api/v1"
        ws_url="wss://api.staging.goplan.io/ws"
        app_url="https://staging.goplan.io"
    fi

    docker build \
        --target production \
        --build-arg NEXT_PUBLIC_API_URL="$api_url" \
        --build-arg NEXT_PUBLIC_WS_URL="$ws_url" \
        --build-arg NEXT_PUBLIC_APP_URL="$app_url" \
        --build-arg NEXT_PUBLIC_APP_NAME="GoPlan" \
        -t "$frontend_image" \
        -f "$PROJECT_ROOT/web/Dockerfile" \
        "$PROJECT_ROOT/web"

    # Push images
    log_info "Pushing images to registry..."
    docker push "$backend_image"
    docker push "$frontend_image"

    log_success "Docker images built and pushed successfully."
}

# Deploy with Helm
deploy() {
    log_info "Deploying to $ENVIRONMENT (namespace: $NAMESPACE)..."

    local helm_cmd="helm upgrade --install goplan $PROJECT_ROOT/deploy/helm/goplan"
    helm_cmd+=" --namespace $NAMESPACE"
    helm_cmd+=" --create-namespace"
    helm_cmd+=" --values $PROJECT_ROOT/deploy/helm/goplan/values.yaml"

    # Add production values if deploying to production
    if [[ "$ENVIRONMENT" == "production" ]]; then
        helm_cmd+=" --values $PROJECT_ROOT/deploy/helm/goplan/values-production.yaml"
    fi

    # Set version
    helm_cmd+=" --set backend.image.tag=$VERSION"
    helm_cmd+=" --set frontend.image.tag=$VERSION"
    helm_cmd+=" --set global.environment=$ENVIRONMENT"
    helm_cmd+=" --set namespace.name=$NAMESPACE"

    # Set environment-specific values
    if [[ "$ENVIRONMENT" == "staging" ]]; then
        helm_cmd+=" --set ingress.hosts.frontend.host=staging.goplan.io"
        helm_cmd+=" --set ingress.hosts.backend.host=api.staging.goplan.io"
        helm_cmd+=" --set frontend.config.apiUrl=https://api.staging.goplan.io/api/v1"
        helm_cmd+=" --set frontend.config.wsUrl=wss://api.staging.goplan.io/ws"
        helm_cmd+=" --set frontend.config.appUrl=https://staging.goplan.io"
    fi

    # Add dry-run flag if specified
    if [[ "$DRY_RUN" == "true" ]]; then
        helm_cmd+=" --dry-run"
        log_warning "DRY RUN MODE - No changes will be made"
    fi

    helm_cmd+=" --wait"
    helm_cmd+=" --timeout 10m"

    # Execute helm command
    log_info "Executing: $helm_cmd"
    eval "$helm_cmd"

    if [[ "$DRY_RUN" != "true" ]]; then
        log_success "Deployment completed successfully."
    fi
}

# Verify deployment
verify() {
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "Skipping verification (dry-run mode)"
        return
    fi

    log_info "Verifying deployment..."

    # Check rollout status
    kubectl rollout status deployment/goplan-backend -n "$NAMESPACE" --timeout=5m
    kubectl rollout status deployment/goplan-frontend -n "$NAMESPACE" --timeout=5m

    # Get pod status
    log_info "Pod status:"
    kubectl get pods -n "$NAMESPACE" -l app.kubernetes.io/name=goplan

    log_success "Deployment verified successfully."
}

# Health check
health_check() {
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "Skipping health check (dry-run mode)"
        return
    fi

    log_info "Running health checks..."

    local api_host frontend_host
    if [[ "$ENVIRONMENT" == "production" ]]; then
        api_host="https://api.goplan.io"
        frontend_host="https://goplan.io"
    else
        api_host="https://api.staging.goplan.io"
        frontend_host="https://staging.goplan.io"
    fi

    # Wait for services to be ready
    sleep 10

    # Check backend health
    if curl -sf "$api_host/health" > /dev/null; then
        log_success "Backend health check passed"
    else
        log_warning "Backend health check failed"
    fi

    # Check frontend health
    if curl -sf "$frontend_host/api/health" > /dev/null; then
        log_success "Frontend health check passed"
    else
        log_warning "Frontend health check failed"
    fi
}

# Main function
main() {
    echo "=========================================="
    echo "  GoPlan Deployment Script"
    echo "=========================================="
    echo

    parse_args "$@"

    log_info "Environment: $ENVIRONMENT"
    log_info "Version: $VERSION"
    log_info "Namespace: $NAMESPACE"
    log_info "Dry Run: $DRY_RUN"
    echo

    check_prerequisites
    build_and_push
    deploy
    verify
    health_check

    echo
    log_success "Deployment process completed!"
}

# Run main function
main "$@"
