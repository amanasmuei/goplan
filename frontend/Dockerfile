# Build stage
FROM node:20-alpine AS builder

# Enable corepack for Yarn
RUN corepack enable

WORKDIR /app

# Copy package files and yarn config
COPY package.json yarn.lock .yarnrc.yml ./
COPY .yarn .yarn

# Install dependencies
RUN yarn install --immutable

# Copy source code
COPY . .

# Build argument for API URL (embedded at build time)
ARG VITE_API_URL=/api/v1
ENV VITE_API_URL=$VITE_API_URL

# Build the application
RUN yarn build

# Production stage
FROM nginx:alpine AS production

# Create non-root user and fix permissions
RUN addgroup -g 1001 -S appgroup && adduser -u 1001 -S appuser -G appgroup && \
    chown -R appuser:appgroup /var/cache/nginx && \
    chown -R appuser:appgroup /var/log/nginx && \
    touch /var/run/nginx.pid && chown appuser:appgroup /var/run/nginx.pid

# Copy built assets
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Fix ownership of html directory
RUN chown -R appuser:appgroup /usr/share/nginx/html

USER appuser

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]
