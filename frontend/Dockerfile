# Stage 1: Build the React application
FROM node:18-alpine as builder

WORKDIR /app

# Copy the package manager configuration first
COPY package*.json ./

# Copy the .npmrc file to configure npm for this build
COPY .npmrc ./

# Install dependencies. The .npmrc file will be automatically used.
RUN npm install --verbose

# Copy the rest of the application source code
COPY . .

# Build the application
RUN npm run build

# Stage 2: Serve the application using Nginx
FROM nginx:stable-alpine

COPY --from=builder /app/build /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]

