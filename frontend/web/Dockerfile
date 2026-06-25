# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Set build-time environment variables
ENV ORIGIN=http://localhost:5173
ENV PUBLIC_API_URL=http://backend:8080
ENV BETTER_AUTH_SECRET=61cd3f27-239d-4f01-9a7b-ec02b417a99a

# Copy package files
COPY package.json package-lock.json* ./

# Install dependencies
RUN npm ci
RUN mkdir -p /data

# Copy source code
COPY . .

# Build the application
RUN mkdir -p /data && npm run build

# Production stage
FROM node:20-alpine

WORKDIR /app

# Create non-root user for security
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001

# Copy built assets
COPY --from=builder --chown=nodejs:nodejs /app/build ./build
COPY --from=builder --chown=nodejs:nodejs /app/package.json ./
COPY --from=builder --chown=nodejs:nodejs /app/node_modules ./node_modules
COPY --from=builder --chown=nodejs:nodejs /app/package-lock.json* ./

# Set environment
ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=5173

# Expose port
EXPOSE 5173

# Run as non-root user
USER nodejs

# Start the application
CMD ["node", "build"]
