FROM node:22-alpine

WORKDIR /app

# Install pnpm globally
RUN npm install -g pnpm

# Copy package files first for better layer caching
COPY package.json pnpm-lock.yaml ./

# Install dependencies
RUN pnpm install --frozen-lockfile

# Copy source code
COPY . .

EXPOSE 3000

# Run dev server with hot reload
# --host 0.0.0.0 allows external connections from Docker host
CMD ["pnpm", "dev", "--host", "0.0.0.0"]
