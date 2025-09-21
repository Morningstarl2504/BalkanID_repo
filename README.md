https://docs.google.com/document/d/1j816X-Oc7nAsObA9jPuFI1iBfqqUdo1e4bOUZnZGCXE/edit?usp=sharing


Secure File Vault

The Secure File Vault is a robust, full-stack application engineered for secure and efficient file management. The backend, built with Go and the Gin framework, features a RESTful API with a PostgreSQL database, and employs a deduplication strategy using SHA-256 content hashing to optimize storage. It incorporates crucial security measures such as role-based access control for administrative functions, per-user rate limiting, and storage quotas. The modern frontend, developed with React and TypeScript, provides a user-friendly interface for file uploads via drag-and-drop, advanced searching, and controlled file sharing with public download links. Containerized with Docker Compose, the project offers a streamlined development environment and a scalable, production-grade solution for a file vault system.



Project Overview

The Secure File Vault is a system designed to provide a secure environment for file storage, search, and sharing. It's built to be scalable and robust, with a focus on good software design and engineering practices.



Core Features Implemented:

File Deduplication: The system detects duplicate file uploads using SHA-256 hashing. Instead of storing duplicate content, it stores a reference to the existing file content, which is managed by a FileContent model. This approach saves disk space and a ReferenceCount is used to track how many files refer to a specific content. Per-user storage savings are also tracked and displayed.

File Uploads: Users can upload single or multiple files with drag-and-drop functionality on the frontend. The application validates the file's content against its declared MIME type to prevent mismatched uploads (e.g., a .docx file renamed as a .jpg).

File Management & Sharing: Users can view a list of all their files with detailed metadata. They can also toggle a public sharing link for their files and track the number of downloads. The system allows only the file owner to delete their files, and deduplicated files are not permanently deleted until all user references are removed.

Search & Filtering: The application supports searching by filename and filtering by MIME type, size range, and date range. The API handlers (backend/internal/handlers/admin.go, backend/internal/services/file.go) are built to handle these filters efficiently.

Rate Limiting & Quotas: The system enforces per-user API rate limits (default is 2 calls per second) and storage quotas (default is 10 MB), returning appropriate error messages when limits are exceeded.

Admin Panel: An admin dashboard provides an overview of system-wide statistics, including total users, files, and storage usage. Admins can view all files in the system and audit logs. The API endpoints are protected by a dedicated AdminMiddleware.



Tech Stack

Backend: Go (Golang) with the Gin framework for the RESTful API  and GORM as the ORM.

Database: PostgreSQL.

Frontend: React.js with TypeScript, built using Create React App and styled with Tailwind CSS. It uses 

lucide-react for icons and recharts for data visualization on the admin panel.

Containerization: Docker Compose is used for local development and testing, providing containers for both the backend and the PostgreSQL database.




Getting Started

Clone the Repository: Ensure you are in your project's root directory.

Start the Database: Run docker-compose up -d postgres to start the PostgreSQL database container. The database name and user are configured in docker-compose.yml.

Backend Setup: Navigate to the backend directory.

To install dependencies, you can run the following command if you are on Windows: go get github.com/gin-gonic/gin go get gorm.io/gorm gorm.io/driver/postgres... or simply run the setup.bat file.


Frontend Setup:

Navigate to the frontend directory.

Run npm install to install dependencies listed in package.json.

Run the Application:

Run docker-compose up --build from the project root to build and run all services.

The frontend will be accessible at http://localhost:3000 and the backend API at http://localhost:8080/api/v1.

User Guide: How to use the Secure File Vault
Welcome to the Secure File Vault! This guide will walk you through the essential features of the application, from creating an account to managing and sharing your files.

1. Account Setup
Registration: When you first open the application, you'll see a sign-in screen. Click "Create one" to register a new account. You'll need to provide a username, email, and password.

Login: After registration, use your email and password to log in.

2. Dashboard and File Management
After logging in, you'll be taken to your dashboard, where you can see your personal storage statistics and manage your files.

Storage Statistics: At the top of your dashboard, you'll see a panel with your storage stats. It shows:

Total Used (Deduplicated): The actual amount of space your files are taking up on the server.

Original Usage: The total size of all your uploaded files, without accounting for deduplication.

Savings: The amount of storage space saved, shown in both bytes and as a percentage.

Uploading Files: The Upload New Files section lets you add files to your vault.

You can either drag and drop files directly into the designated area or click the area to select files from your computer.

The system supports uploading multiple files at once.

Managing Your Files: The Manage Files section displays a list of all your uploaded files. Each file entry shows its filename, size, uploader, upload date, and download count.

Search: Use the search bar to find files by their name.

Filter: Click the filter icon to show advanced filtering options. You can filter files by MIME type (e.g., Image, PDF), a specific size range, or a date range.

3. File Actions
For each file in your list, you have several options:

Share: Click the share icon to toggle the public sharing status of a file. When enabled, a public link is generated that anyone can use to download the file without logging in.

Delete: Click the trash can icon to permanently delete a file. A confirmation prompt will appear to prevent accidental deletion.

Download: Click the download icon to download your file.

4. Admin Dashboard (for Admin users only)
If you have an admin account, you'll see a different dashboard with system-wide information:

System Overview: This section shows key metrics like the total number of users and files, and overall storage deduplication statistics.

All Files in System: This list shows every file that has been uploaded by all users on the platform.