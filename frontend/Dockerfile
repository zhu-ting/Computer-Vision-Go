# ── Stage 1: Build the React app ──────────────────────
FROM node:20-alpine AS builder

WORKDIR /app

# Cache npm dependencies separately from source code
COPY package.json package-lock.json* ./
RUN npm ci

COPY . .
RUN npm run build

# ── Stage 2: Serve with nginx ───────────────────────────
FROM nginx:1.25-alpine

# Replace the default nginx config to enable SPA routing
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built assets from the builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
