import React, { useState, useCallback, useEffect } from 'react';
import { 
    UploadCloud, 
    File, 
    Folder, 
    Loader2, 
    CheckCircle, 
    AlertTriangle, 
    User, 
    LogIn, 
    UserPlus,
    LogOut,
    Settings
} from 'lucide-react';

// --- TYPE DEFINITIONS ---
interface FileItem {
    id: number;
    filename: string;
    original_filename: string;
    content: {
        file_size: number;
        mime_type: string;
    };
    owner: {
        username: string;
    };
    uploaded_at: string;
    download_count: number;
    is_public: boolean;
}

interface FolderItem { 
    id: string; 
    name: string; 
    files: FileItem[]; 
}

interface User {
    id: number;
    username: string;
    email: string;
    is_admin: boolean;
}

// --- API HELPER ---
const API_BASE_URL = 'http://localhost:8080/api/v1';

const getAuthToken = () => {
    return localStorage.getItem('token') || localStorage.getItem('authToken');
};

const apiRequest = async (endpoint: string, options: RequestInit = {}) => {
    const token = getAuthToken();
    const headers: Record<string, string> = {
        ...options.headers as Record<string, string>,
    };

    // Only set Content-Type if not uploading files
    if (!(options.body instanceof FormData)) {
        headers['Content-Type'] = 'application/json';
    }

    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    try {
        const response = await fetch(`${API_BASE_URL}${endpoint}`, {
            ...options,
            headers,
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
            throw new Error(errorData.error || `HTTP ${response.status}`);
        }

        return response.json();
    } catch (error) {
        if (error instanceof TypeError && error.message.includes('fetch')) {
            throw new Error('Cannot connect to server. Please check if the backend is running on http://localhost:8080');
        }
        throw error;
    }
};

// --- AUTH COMPONENTS ---
const AuthForm: React.FC<{ onAuthSuccess: (user: User) => void }> = ({ onAuthSuccess }) => {
    const [isLogin, setIsLogin] = useState(true);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [formData, setFormData] = useState({
        username: '',
        email: '',
        password: ''
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsLoading(true);
        setError(null);

        try {
            const endpoint = isLogin ? '/auth/login' : '/auth/register';
            const payload = isLogin 
                ? { email: formData.email, password: formData.password }
                : formData;

            const response = await fetch(`${API_BASE_URL}${endpoint}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload),
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Authentication failed');
            }

            const data = await response.json();
            localStorage.setItem('token', data.data.token);
            onAuthSuccess(data.data.user);
        } catch (err: any) {
            if (err.message.includes('fetch')) {
                setError('Cannot connect to server. Please check if the backend is running.');
            } else {
                setError(err.message);
            }
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-gray-50 flex items-center justify-center">
            <div className="bg-white p-8 rounded-lg shadow-md w-full max-w-md">
                <h2 className="text-2xl font-bold text-center mb-6">
                    {isLogin ? 'Sign In' : 'Create Account'}
                </h2>

                <form onSubmit={handleSubmit} className="space-y-4">
                    {!isLogin && (
                        <input
                            type="text"
                            placeholder="Username"
                            value={formData.username}
                            onChange={(e) => setFormData(prev => ({ ...prev, username: e.target.value }))}
                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                            required
                        />
                    )}
                    
                    <input
                        type="email"
                        placeholder="Email"
                        value={formData.email}
                        onChange={(e) => setFormData(prev => ({ ...prev, email: e.target.value }))}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        required
                    />
                    
                    <input
                        type="password"
                        placeholder="Password"
                        value={formData.password}
                        onChange={(e) => setFormData(prev => ({ ...prev, password: e.target.value }))}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        required
                        minLength={6}
                    />

                    <button
                        type="submit"
                        disabled={isLoading}
                        className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 disabled:opacity-50 flex items-center justify-center"
                    >
                        {isLoading ? (
                            <Loader2 className="w-4 h-4 animate-spin mr-2" />
                        ) : isLogin ? (
                            <LogIn className="w-4 h-4 mr-2" />
                        ) : (
                            <UserPlus className="w-4 h-4 mr-2" />
                        )}
                        {isLoading ? 'Please wait...' : (isLogin ? 'Sign In' : 'Create Account')}
                    </button>
                </form>

                {error && (
                    <div className="mt-4 text-red-700 bg-red-100 p-3 rounded-lg flex items-center">
                        <AlertTriangle className="w-5 h-5 mr-2" />
                        {error}
                    </div>
                )}

                <p className="mt-6 text-center text-gray-600">
                    {isLogin ? "Don't have an account?" : "Already have an account?"}
                    <button
                        type="button"
                        onClick={() => setIsLogin(!isLogin)}
                        className="ml-2 text-blue-600 hover:text-blue-800 font-medium"
                    >
                        {isLogin ? 'Create one' : 'Sign in'}
                    </button>
                </p>
            </div>
        </div>
    );
};

// --- UPLOAD COMPONENT ---
const UploadZone: React.FC<{ onUploadSuccess: () => void }> = ({ onUploadSuccess }) => {
    const [isUploading, setIsUploading] = useState(false);
    const [uploadError, setUploadError] = useState<string | null>(null);
    const [uploadSuccess, setUploadSuccess] = useState(false);

    const uploadFiles = async (files: File[]) => {
        setIsUploading(true);
        setUploadError(null);
        setUploadSuccess(false);

        const formData = new FormData();
        files.forEach(file => formData.append('files', file));

        try {
            const token = getAuthToken();
            if (!token) {
                throw new Error('Please log in to upload files');
            }

            const response = await fetch(`${API_BASE_URL}/files/upload`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`
                },
                body: formData,
            });

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({ error: 'Upload failed' }));
                throw new Error(errorData.error || `Upload failed: ${response.status}`);
            }

            const result = await response.json();
            setUploadSuccess(true);
            onUploadSuccess();
            
            setTimeout(() => setUploadSuccess(false), 5000);
        } catch (error: any) {
            if (error.message.includes('fetch') || error.name === 'TypeError') {
                setUploadError('Network Error: Cannot connect to backend. Please check if the server is running on http://localhost:8080');
            } else {
                setUploadError(error.message);
            }
        } finally {
            setIsUploading(false);
        }
    };

    const onDrop = useCallback((acceptedFiles: File[]) => {
        if (acceptedFiles.length > 0) {
            uploadFiles(acceptedFiles);
        }
    }, []);

    // Simple drag and drop implementation
    const [isDragActive, setIsDragActive] = useState(false);

    const handleDragEnter = (e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setIsDragActive(true);
    };

    const handleDragLeave = (e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setIsDragActive(false);
    };

    const handleDragOver = (e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
    };

    const handleDrop = (e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setIsDragActive(false);
        
        const files = Array.from(e.dataTransfer.files);
        if (files.length > 0 && !isUploading) {
            uploadFiles(files);
        }
    };

    const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
        const files = Array.from(e.target.files || []);
        if (files.length > 0) {
            uploadFiles(files);
        }
    };

    return (
        <div>
            <div 
                onDragEnter={handleDragEnter}
                onDragLeave={handleDragLeave}
                onDragOver={handleDragOver}
                onDrop={handleDrop}
                onClick={() => !isUploading && document.getElementById('file-upload')?.click()}
                className={`relative border-2 border-dashed rounded-lg p-12 text-center transition-colors cursor-pointer
                    ${isUploading ? 'bg-gray-100 cursor-not-allowed' : 'hover:border-blue-400'} 
                    ${isDragActive ? 'border-blue-500 bg-blue-50' : 'border-gray-300'}`}
            >
                <input 
                    type="file" 
                    id="file-upload" 
                    name="files" 
                    multiple 
                    onChange={handleFileSelect}
                    className="hidden"
                    disabled={isUploading}
                />
                {isUploading ? (
                    <>
                        <Loader2 className="w-12 h-12 text-blue-500 animate-spin mx-auto mb-4" />
                        <p className="text-gray-600">Uploading files...</p>
                    </>
                ) : (
                    <>
                        <UploadCloud className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                        <p className="text-lg text-gray-700 mb-2">
                            {isDragActive ? 'Drop the files here...' : 'Drag & drop some files here, or click to select files'}
                        </p>
                        <p className="text-sm text-gray-500">Single or multiple file upload supported</p>
                    </>
                )}
            </div>

            {uploadSuccess && (
                <div className="mt-4 text-green-700 bg-green-100 p-3 rounded-lg flex items-center">
                    <CheckCircle className="w-5 h-5 mr-2 flex-shrink-0" />
                    Files uploaded successfully!
                </div>
            )}

            {uploadError && (
                <div className="mt-4 text-red-700 bg-red-100 p-3 rounded-lg flex items-center">
                    <AlertTriangle className="w-5 h-5 mr-2 flex-shrink-0" />
                    {uploadError}
                </div>
            )}
        </div>
    );
};

// --- FILE LIST COMPONENT ---
const FileList: React.FC<{ files: FileItem[], folders: FolderItem[] }> = ({ files, folders }) => {
    const formatFileSize = (bytes: number) => {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleDateString();
    };

    if (files.length === 0 && folders.length === 0) {
        return (
            <div className="text-center py-12 text-gray-500">
                                <File className="w-16 h-16 mx-auto mb-4 text-gray-300" />
                <p className="text-lg">No files uploaded yet</p>
                <p className="text-sm">Upload your first file to get started</p>
            </div>
        );
    }

    return (
        <div>
            <div className="flex items-center justify-between mb-6">
                <h3 className="text-xl font-semibold">Your Files</h3>
                <div className="text-sm text-gray-600">
                    {files.length} files, {folders.length} folders
                </div>
            </div>

            {/* Folders */}
            {folders.length > 0 && (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
                    {folders.map((folder) => (
                        <div key={folder.id} className="bg-white p-4 rounded-lg border hover:shadow-md transition-shadow cursor-pointer">
                            <div className="flex items-center mb-2">
                                <Folder className="w-5 h-5 text-yellow-500 mr-2" />
                                <span className="font-medium truncate">{folder.name}</span>
                            </div>
                            <p className="text-sm text-gray-600">{folder.files.length} files</p>
                        </div>
                    ))}
                </div>
            )}

            {/* Files */}
            <div className="space-y-3">
                {files.map((file) => (
                    <div key={file.id} className="bg-white p-4 rounded-lg border hover:shadow-md transition-shadow">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center flex-1 min-w-0">
                                <File className="w-5 h-5 text-blue-500 mr-3 flex-shrink-0" />
                                <div className="min-w-0 flex-1">
                                    <p className="font-medium text-gray-900 truncate">
                                        {file.original_filename}
                                    </p>
                                    <div className="text-sm text-gray-500 flex items-center space-x-4">
                                        <span>{formatFileSize(file.content.file_size)}</span>
                                        <span>by {file.owner.username}</span>
                                        <span>{formatDate(file.uploaded_at)}</span>
                                        <span>{file.download_count} downloads</span>
                                        {file.is_public && (
                                            <span className="px-2 py-1 bg-green-100 text-green-800 rounded-full text-xs">
                                                Public
                                            </span>
                                        )}
                                    </div>
                                </div>
                            </div>
                            <div className="flex items-center space-x-2 ml-4">
                                <button 
                                    className="text-blue-600 hover:text-blue-800 text-sm"
                                    onClick={() => {
                                        const token = getAuthToken();
                                        const url = `${API_BASE_URL}/files/${file.id}/download`;
                                        window.open(url + (token ? `?token=${token}` : ''), '_blank');
                                    }}
                                >
                                    Download
                                </button>
                            </div>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

// --- MAIN APP COMPONENT ---
function App() {
    const [user, setUser] = useState<User | null>(null);
    const [files, setFiles] = useState<FileItem[]>([]);
    const [folders, setFolders] = useState<FolderItem[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const checkAuthStatus = useCallback(async () => {
        const token = getAuthToken();
        if (!token) {
            setIsLoading(false);
            return;
        }

        try {
            const userData = await apiRequest('/profile');
            setUser(userData.data);
        } catch (err: any) {
            console.error('Auth check failed:', err);
            localStorage.removeItem('token');
            localStorage.removeItem('authToken');
        } finally {
            setIsLoading(false);
        }
    }, []);

    const fetchFiles = useCallback(async () => {
        if (!user) return;
        
        setError(null);
        try {
            const data = await apiRequest('/files');
            setFiles(data.data.files || []);
            setFolders([]);  // Implement folder fetching if needed
        } catch (err: any) {
            setError(err.message);
        }
    }, [user]);

    const handleLogout = () => {
        localStorage.removeItem('token');
        localStorage.removeItem('authToken');
        setUser(null);
        setFiles([]);
        setFolders([]);
    };

    const handleAuthSuccess = (userData: User) => {
        setUser(userData);
    };

    useEffect(() => {
        checkAuthStatus();
    }, [checkAuthStatus]);

    useEffect(() => {
        if (user) {
            fetchFiles();
        }
    }, [user, fetchFiles]);

    if (isLoading) {
        return (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center">
                <div className="text-center">
                    <Loader2 className="w-8 h-8 animate-spin text-blue-600 mx-auto mb-4" />
                    <p className="text-gray-600">Connecting to server...</p>
                </div>
            </div>
        );
    }

    if (!user) {
        return <AuthForm onAuthSuccess={handleAuthSuccess} />;
    }

    return (
        <div className="bg-gray-50 min-h-screen">
            <header className="bg-white shadow-sm border-b">
                <div className="container mx-auto px-4 py-4">
                    <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-3">
                            <File className="w-8 h-8 text-blue-600" />
                            <h1 className="text-2xl font-bold text-gray-900">Secure File Vault</h1>
                        </div>
                        <div className="flex items-center space-x-4">
                            <div className="flex items-center space-x-2 text-gray-700">
                                <User className="w-4 h-4" />
                                <span>{user.username}</span>
                                {user.is_admin && (
                                    <span className="px-2 py-1 bg-blue-100 text-blue-800 rounded-full text-xs">
                                        Admin
                                    </span>
                                )}
                            </div>
                            <button
                                onClick={handleLogout}
                                className="flex items-center space-x-1 text-gray-600 hover:text-gray-800 px-3 py-2 rounded-md hover:bg-gray-100"
                            >
                                <LogOut className="w-4 h-4" />
                                <span>Logout</span>
                            </button>
                        </div>
                    </div>
                </div>
            </header>

            <main className="container mx-auto px-4 py-8">
                <div className="bg-white rounded-lg shadow-sm border p-6 mb-8">
                    <h2 className="text-xl font-semibold mb-4 flex items-center">
                        <UploadCloud className="w-5 h-5 mr-2" />
                        Upload New Files
                    </h2>
                    <UploadZone onUploadSuccess={fetchFiles} />
                </div>

                <div className="bg-white rounded-lg shadow-sm border p-6">
                    {error && (
                        <div className="mb-6 text-red-700 bg-red-100 p-4 rounded-lg flex items-center">
                            <AlertTriangle className="w-5 h-5 mr-2 flex-shrink-0" />
                            {error}
                        </div>
                    )}
                    <FileList files={files} folders={folders} />
                </div>
            </main>
        </div>
    );
}

export default App;