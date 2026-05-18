# --- Stage 1: Installation & compilation ---
FROM node:22-alpine AS builder

WORKDIR /app

# Cache package dependencies
COPY package.json package-lock.json ./
RUN npm ci --legacy-peer-deps

# Copy source trees
COPY . .

# Compile optimized static output
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build

# --- Stage 2: Production runtime runner ---
FROM node:22-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

# Copy Next.js compilation layers
COPY --from=builder /app/package.json ./
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public

EXPOSE 3000

CMD ["npm", "start"]
