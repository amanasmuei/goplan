#!/usr/bin/env bash
# ===========================================
# GoPlan Rollback Script
# Rollback to previous deployment version
# ===========================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT="${ENVIRONMENT:-staging}"
REVISION="${REVISION:-}"
NAMESPACE=""
DRY_RUN="${DRY_RUN:-false}"
LIST_REVISIONS="${LIST_REVISIONS:-false}"

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

Rollback GoPlan deployment to a previous version.

OPTIONS:
    -e, --environment   Environment to rollback (staging|production) [default: staging]
    -r, --revision      Helm revision to rollback to [default: previous revision]
    -n, --namespace     Kubernetes namespace [default: goplan-{environment}]
    -l, --list          List available revisions
    -d, --dry-run       Perform a dry run without making changes
    -h, --help          Show this help message

EXAMPLES:
    $(basename "$0") -e staging                    # Rollback to previous version
    $(basename "$0") -e production -r 5            # Rollback to revision 5
    $(basename "$0") -e staging --list             # List available revisions
    $(basename "$0") -e production --dry-run       # Dry run rollback

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
            -r|--revision)
                REVISION="$2"
                shift 2
                ;;
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            -l|--list)
                LIST_REVISIONS="true"
                shift
                ;;
            -d|--dry-run)
                DRY_RUN="true"
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

# List available revisions
list_revisions() {
    log_info "Listing available revisions for goplan in namespace: $NAMESPACE"
    echo

    helm history goplan -n "$NAMESPACE" --max 20

    echo
    log_info "Use -r <revision> to rollback to a specific revision"
}

# Get current revision
get_current_revision() {
    helm history goplan -n "$NAMESPACE" --max 1 -o json | grep -o '"revision":[0-9]*' | cut -d: -f2
}

# Rollback deployment
rollback() {
    local current_revision
    current_revision=$(get_current_revision)

    # If no revision specified, rollback to previous
    if [[ -z "$REVISION" ]]; then
        REVISION=$((current_revision - 1))
        if [[ $REVISION -lt 1 ]]; then
            log_error "No previous revision available for rollback."
            exit 1
        fi
    fi

    log_info "Current revision: $current_revision"
    log_info "Rolling back to revision: $REVISION"

    # Production confirmation
    if [[ "$ENVIRONMENT" == "production" && "$DRY_RUN" != "true" ]]; then
        log_warning "You are about to rollback PRODUCTION!"
        read -p "Are you sure you want to continue? (yes/no): " confirm
        if [[ "$confirm" != "yes" ]]; then
            log_info "Rollback cancelled."
            exit 0
        fi
    fi

    local helm_cmd="helm rollback goplan $REVISION -n $NAMESPACE"

    if [[ "$DRY_RUN" == "true" ]]; then
        helm_cmd+=" --dry-run"
        log_warning "DRY RUN MODE - No changes will be made"
    fi

    # Execute rollback
    log_info "Executing: $helm_cmd"
    eval "$helm_cmd"

    if [[ "$DRY_RUN" != "true" ]]; then
        log_success "Rollback to revision $REVISION completed."
    fi
}

# Verify rollback
verify() {
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "Skipping verification (dry-run mode)"
        return
    fi

    log_info "Verifying rollback..."

    # Check rollout status
    kubectl rollout status deployment/goplan-backend -n "$NAMESPACE" --timeout=5m
    kubectl rollout status deployment/goplan-frontend -n "$NAMESPACE" --timeout=5m

    # Get pod status
    log_info "Pod status:"
    kubectl get pods -n "$NAMESPACE" -l app.kubernetes.io/name=goplan

    # Show current helm status
    log_info "Current Helm release status:"
    helm status goplan -n "$NAMESPACE"

    log_success "Rollback verified successfully."
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
        log_error "Backend health check failed"
        exit 1
    fi

    # Check frontend health
    if curl -sf "$frontend_host/api/health" > /dev/null; then
        log_success "Frontend health check passed"
    else
        log_error "Frontend health check failed"
        exit 1
    fi
}

# Main function
main() {
    echo "=========================================="
    echo "  GoPlan Rollback Script"
    echo "=========================================="
    echo

    parse_args "$@"

    log_info "Environment: $ENVIRONMENT"
    log_info "Namespace: $NAMESPACE"
    echo

    check_prerequisites

    if [[ "$LIST_REVISIONS" == "true" ]]; then
        list_revisions
        exit 0
    fi

    rollback
    verify
    health_check

    echo
    log_success "Rollback process completed!"
}

# Run main function
main "$@"
