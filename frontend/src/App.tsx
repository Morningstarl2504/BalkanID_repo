// morningstarl2504/balkanid_repo/BalkanID_repo-f1fc3ed153144eb6d79e3c90f73a0f3d312b9c79/frontend/src/App.tsx
import React, { useState, useCallback, useEffect } from 'react';
import {
    UploadCloud, File as FileIcon, Loader2, CheckCircle, AlertTriangle,
    User, LogIn, UserPlus, LogOut, ShieldCheck, Trash2, Download, Search, Filter
} from 'lucide-react';
import { PieChart, Pie, Cell, Tooltip, Legend, ResponsiveContainer } from 'recharts';

// --- TYPE DEFINITIONS ---
interface FileContent {
    file_size: number;
    mime_type: string;
}
interface FileOwner {
    username: string;
}
interface FileItem {
    id: number;
    original_filename: string;
    content: FileContent;
    user: FileOwner;
    created_at: string;
    download_count: number;
    is_public: boolean;
}
interface User {
    id: number;
    username: string;
    email: string;
    is_admin: boolean;
}
interface StorageStats {
    total_storage_used: number;
    original_storage_usage: number;
    storage_savings_bytes: number;
    storage_savings_percentage: number;
    user_quota: number;
}

// --- API HELPER ---
const API_BASE_URL = 'http://localhost:8080/api/v1';
const getAuthToken = () => localStorage.getItem('token');

const apiRequest = async (endpoint: string, options: RequestInit = {}) => {
    const token = getAuthToken();
    const headers: Record<string, string> = { ...(options.headers as Record<string, string>) };
    if (!(options.body instanceof FormData)) {
        headers['Content-Type'] = 'application/json';
    }
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }
    try {
        const response = await fetch(`${API_BASE_URL}${endpoint}`, { ...options, headers });
        if (!response.ok) {
            if (response.status === 401) {
                localStorage.removeItem('token');
                window.location.reload();
            }
            const errorData = await response.json().catch(() => ({ error: 'An unknown error occurred' }));
            throw new Error(errorData.error || `Request failed with status ${response.status}`);
        }
        if (response.status === 204) return null;
        return response.json();
    } catch (error) {
        if (error instanceof TypeError) {
            throw new Error('Cannot connect to server. Is the backend running?');
        }
        throw error;
    }
};

// --- HELPER FUNCTIONS ---
const formatBytes = (bytes: number, decimals = 2) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
};

// --- UI COMPONENTS ---
const AuthForm: React.FC<{ onAuthSuccess: (user: User) => void }> = ({ onAuthSuccess }) => {
    const [isLogin, setIsLogin] = useState(true);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [formData, setFormData] = useState({ username: '', email: '', password: '' });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsLoading(true);
        setError(null);
        try {
            const endpoint = isLogin ? '/auth/login' : '/auth/register';
            const payload = isLogin ? { email: formData.email, password: formData.password } : formData;
            const data = await apiRequest(endpoint, { method: 'POST', body: JSON.stringify(payload) });
            localStorage.setItem('token', data.data.token);
            onAuthSuccess(data.data.user);
        } catch (err: any) {
            setError(err.message);
        } finally {
            setIsLoading(false);
        }
    };
    return (
        <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
            <div className="bg-white p-8 rounded-xl shadow-md w-full max-w-md">
                <div className="flex justify-center items-center mb-6 space-x-3">
                    <ShieldCheck className="w-10 h-10 text-blue-600" />
                    <h1 className="text-3xl font-bold text-gray-800">Secure File Vault</h1>
                </div>
                <h2 className="text-2xl font-bold text-center mb-6 text-gray-700">{isLogin ? 'Sign In' : 'Create Account'}</h2>
                <form onSubmit={handleSubmit} className="space-y-4">
                    {!isLogin && (<input type="text" placeholder="Username" value={formData.username} onChange={(e) => setFormData(p => ({ ...p, username: e.target.value }))} className="w-full px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500" required />)}
                    <input type="email" placeholder="Email" value={formData.email} onChange={(e) => setFormData(p => ({ ...p, email: e.target.value }))} className="w-full px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500" required />
                    <input type="password" placeholder="Password" value={formData.password} onChange={(e) => setFormData(p => ({ ...p, password: e.target.value }))} className="w-full px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500" required minLength={6} />
                    <button type="submit" disabled={isLoading} className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 disabled:opacity-50 flex items-center justify-center transition-colors">
                        {isLoading ? <Loader2 className="w-5 h-5 animate-spin mr-2" /> : isLogin ? <LogIn className="w-5 h-5 mr-2" /> : <UserPlus className="w-5 h-5 mr-2" />}
                        {isLoading ? 'Please wait...' : (isLogin ? 'Sign In' : 'Create Account')}
                    </button>
                </form>
                {error && (<div className="mt-4 text-red-700 bg-red-100 p-3 rounded-lg flex items-center"><AlertTriangle className="w-5 h-5 mr-2" /> {error}</div>)}
                <p className="mt-6 text-center text-gray-600">
                    {isLogin ? "Don't have an account?" : "Already have an account?"}
                    <button type="button" onClick={() => setIsLogin(!isLogin)} className="ml-2 text-blue-600 hover:text-blue-800 font-medium">
                        {isLogin ? 'Create one' : 'Sign in'}
                    </button>
                </p>
            </div>
        </div>
    );
};

const Notification: React.FC<{ message: string; type: 'success' | 'error' }> = ({ message, type }) => {
    const bgColor = type === 'success' ? 'bg-green-100' : 'bg-red-100';
    const textColor = type === 'success' ? 'text-green-800' : 'text-red-800';
    const Icon = type === 'success' ? CheckCircle : AlertTriangle;
    return (
        <div className={`fixed top-5 right-5 z-50 p-4 rounded-lg shadow-lg flex items-center ${bgColor} ${textColor}`}>
            <Icon className="w-5 h-5 mr-3" />
            <p>{message}</p>
        </div>
    );
};

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
            await apiRequest('/files/upload', { method: 'POST', body: formData });
            setUploadSuccess(true);
            onUploadSuccess();
            setTimeout(() => setUploadSuccess(false), 5000);
        } catch (error: any) {
            setUploadError(error.message);
        } finally {
            setIsUploading(false);
        }
    };
    const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
        const files = Array.from(e.target.files || []);
        if (files.length > 0) uploadFiles(files);
    };
    return (
        <div>
            <label htmlFor="file-upload" className={`relative border-2 border-dashed rounded-lg p-12 text-center transition-colors cursor-pointer block ${isUploading ? 'bg-gray-100 cursor-not-allowed' : 'hover:border-blue-400'}`}>
                <input type="file" id="file-upload" name="files" multiple onChange={handleFileSelect} className="hidden" disabled={isUploading} />
                {isUploading ? (<><Loader2 className="w-12 h-12 text-blue-500 animate-spin mx-auto mb-4" /><p className="text-gray-600">Uploading files...</p></>) : (<><UploadCloud className="w-12 h-12 text-gray-400 mx-auto mb-4" /><p className="text-lg text-gray-700 mb-2">Drag & drop files or click to select</p><p className="text-sm text-gray-500">Multiple file upload supported</p></>)}
            </label>
            {uploadSuccess && (<div className="mt-4 text-green-700 bg-green-100 p-3 rounded-lg flex items-center"><CheckCircle className="w-5 h-5 mr-2" />Files uploaded successfully!</div>)}
            {uploadError && (<div className="mt-4 text-red-700 bg-red-100 p-3 rounded-lg flex items-center"><AlertTriangle className="w-5 h-5 mr-2" />{uploadError}</div>)}
        </div>
    );
};

const FileList: React.FC<{ files: FileItem[], onDelete: (fileId: number) => void }> = ({ files, onDelete }) => {
    if (files.length === 0) {
        return <div className="text-center py-12 text-gray-500"><FileIcon className="w-16 h-16 mx-auto mb-4 text-gray-300" /><p>No files found.</p></div>;
    }
    return (
        <div className="space-y-3">
            {files.map((file) => (
                <div key={file.id} className="bg-white p-4 rounded-lg border hover:shadow-md transition-shadow flex items-center justify-between">
                    <div className="flex items-center flex-1 min-w-0">
                        <FileIcon className="w-6 h-6 text-blue-500 mr-4 flex-shrink-0" />
                        <div className="min-w-0 flex-1">
                            <p className="font-medium text-gray-900 truncate">{file.original_filename}</p>
                            <div className="text-sm text-gray-500 flex items-center flex-wrap gap-x-4 gap-y-1 mt-1">
                                <span>{formatBytes(file.content.file_size)}</span>
                                <span>by {file.user.username}</span>
                                <span>{new Date(file.created_at).toLocaleDateString()}</span>
                                <span className="flex items-center"><Download className="w-4 h-4 mr-1" /> {file.download_count}</span>
                            </div>
                        </div>
                    </div>
                    <div className="flex items-center space-x-2 ml-4">
                        <button onClick={() => onDelete(file.id)} className="p-2 text-gray-500 hover:text-red-600 rounded-full hover:bg-gray-100 transition-colors"><Trash2 size={18} /></button>
                        <a href={`${API_BASE_URL}/files/${file.id}/download`} target="_blank" rel="noopener noreferrer" className="p-2 text-gray-500 hover:text-green-600 rounded-full hover:bg-gray-100 transition-colors"><Download size={18} /></a>
                    </div>
                </div>
            ))}
        </div>
    );
};

const StorageStatsDisplay: React.FC<{ stats: StorageStats | null }> = ({ stats }) => {
    if (!stats) {
        return (
            <div className="bg-white rounded-lg shadow-sm border p-6">
                <h2 className="text-xl font-semibold mb-4">Storage Statistics</h2>
                <div className="flex items-center justify-center h-40">
                    <Loader2 className="w-6 h-6 animate-spin text-gray-400" />
                    <p className="ml-3 text-gray-500">Loading statistics...</p>
                </div>
            </div>
        );
    }
    const quotaUsedPercentage = stats.user_quota > 0 ? (stats.total_storage_used / stats.user_quota) * 100 : 0;
    const pieData = [
        { name: 'Used', value: stats.total_storage_used },
        { name: 'Free', value: Math.max(0, stats.user_quota - stats.total_storage_used) },
    ];
    const COLORS = ['#3B82F6', '#E5E7EB'];

    return (
        <div className="bg-white rounded-lg shadow-sm border p-6">
            <h2 className="text-xl font-semibold mb-4">Storage Statistics</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-center">
                <div>
                    <p className="text-sm text-gray-600">Total Used (Deduplicated): <span className="font-bold text-blue-600">{formatBytes(stats.total_storage_used)}</span></p>
                    <p className="text-sm text-gray-600">Original Usage: <span className="font-bold text-gray-800">{formatBytes(stats.original_storage_usage)}</span></p>
                    <p className="text-sm text-green-600 font-semibold">Savings: {formatBytes(stats.storage_savings_bytes)} ({stats.storage_savings_percentage.toFixed(2)}%)</p>
                    <div className="w-full bg-gray-200 rounded-full h-2.5 my-4">
                        <div className="bg-blue-600 h-2.5 rounded-full" style={{ width: `${quotaUsedPercentage}%` }}></div>
                    </div>
                    <p className="text-xs text-gray-500 text-right">{formatBytes(stats.total_storage_used)} of {formatBytes(stats.user_quota)} used</p>
                </div>
                <div className="h-40">
                    <ResponsiveContainer>
                        <PieChart>
                            <Pie data={pieData} dataKey="value" nameKey="name" cx="50%" cy="50%" outerRadius={60} fill="#8884d8">
                                {pieData.map((entry, index) => <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />)}
                            </Pie>
                            <Tooltip formatter={(value: number) => formatBytes(value)} />
                            <Legend />
                        </PieChart>
                    </ResponsiveContainer>
                </div>
            </div>
        </div>
    );
};

const SearchAndFilter: React.FC<{ onFilterChange: (filters: any) => void }> = ({ onFilterChange }) => {
    const [filters, setFilters] = useState({
        filename: '',
        mime_type: '',
        min_size: '',
        max_size: '',
        start_date: '',
        end_date: ''
    });
    const [showAdvanced, setShowAdvanced] = useState(false);

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
        setFilters({ ...filters, [e.target.name]: e.target.value });
    };

    const handleApplyFilters = () => {
        const activeFilters: { [key: string]: any } = {};
        // Don't include empty strings in the filter object
        for (const [key, value] of Object.entries(filters)) {
            if (value) {
                activeFilters[key] = value;
            }
        }
        onFilterChange(activeFilters);
    };
    
    return (
        <div className="bg-gray-50 p-4 rounded-lg border mb-6">
            <div className="flex gap-2">
                <input 
                    type="text" 
                    name="filename"
                    value={filters.filename}
                    onChange={handleInputChange}
                    placeholder="Search by filename..."
                    className="w-full p-2 border rounded-md"
                />
                <button onClick={() => setShowAdvanced(!showAdvanced)} className="p-2 border rounded-md hover:bg-gray-100" title="Advanced Filters">
                    <Filter className="w-5 h-5 text-gray-600" />
                </button>
                <button onClick={handleApplyFilters} className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700">
                    <Search className="w-5 h-5" />
                </button>
            </div>
            {showAdvanced && (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mt-4 border-t pt-4">
                    <select name="mime_type" value={filters.mime_type} onChange={handleInputChange} className="p-2 border rounded-md bg-white">
                        <option value="">All Types</option>
                        <option value="image">Image (jpg, png, etc)</option>
                        <option value="application/pdf">PDF</option>
                        <option value="application/vnd.openxmlformats-officedocument">Office Document (docx, xlsx)</option>
                        <option value="text">Text (txt, csv, etc)</option>
                        <option value="application/zip">Zip Archive</option>
                    </select>
                    <input type="number" name="min_size" value={filters.min_size} onChange={handleInputChange} placeholder="Min Size (Bytes)" className="p-2 border rounded-md" />
                    <input type="number" name="max_size" value={filters.max_size} onChange={handleInputChange} placeholder="Max Size (Bytes)" className="p-2 border rounded-md" />
                    <div/>
                    <label className="flex items-center gap-2 text-sm text-gray-600">From: <input type="date" name="start_date" value={filters.start_date} onChange={handleInputChange} className="p-2 border rounded-md w-full" /></label>
                    <label className="flex items-center gap-2 text-sm text-gray-600">To: <input type="date" name="end_date" value={filters.end_date} onChange={handleInputChange} className="p-2 border rounded-md w-full" /></label>
                </div>
            )}
        </div>
    );
};


// --- MAIN APP COMPONENT ---
function App() {
    const [user, setUser] = useState<User | null>(null);
    const [files, setFiles] = useState<FileItem[]>([]);
    const [storageStats, setStorageStats] = useState<StorageStats | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [notification, setNotification] = useState<{ message: string, type: 'success' | 'error' } | null>(null);
    const [activeFilters, setActiveFilters] = useState({});

    const showNotification = (message: string, type: 'success' | 'error') => {
        setNotification({ message, type });
        setTimeout(() => setNotification(null), 4000);
    };

    const checkAuthStatus = useCallback(async () => {
        setIsLoading(true);
        const token = getAuthToken();
        if (!token) {
            setIsLoading(false);
            return;
        }
        try {
            const profileData = await apiRequest('/profile');
            setUser(profileData.data);
        } catch (err) {
            localStorage.removeItem('token');
        } finally {
            setIsLoading(false);
        }
    }, []);

    const fetchData = useCallback(async () => {
        if (!user) return;
        setError(null);
        try {
            const defaultParams = { page: 1, limit: 100 };
            const params = new URLSearchParams({ ...defaultParams, ...activeFilters } as any).toString();
            
            const [filesData, statsData] = await Promise.all([
                apiRequest(`/files?${params}`),
                apiRequest('/storage/stats')
            ]);
            setFiles(filesData.data.files || []);
            setStorageStats(statsData.data);
        } catch (err: any) {
            setError(err.message);
        }
    }, [user, activeFilters]);

    useEffect(() => {
        checkAuthStatus();
    }, [checkAuthStatus]);

    useEffect(() => {
        if (user) {
            fetchData();
        }
    }, [user, fetchData]);

    const handleLogout = () => {
        localStorage.removeItem('token');
        setUser(null);
        setFiles([]);
        setStorageStats(null);
    };
    
    const handleDeleteFile = async (fileId: number) => {
        if (window.confirm('Are you sure you want to delete this file?')) {
            try {
                await apiRequest(`/files/${fileId}`, { method: 'DELETE' });
                showNotification('File deleted successfully.', 'success');
                fetchData();
            } catch (err: any) {
                showNotification(`Error: ${err.message}`, 'error');
            }
        }
    };

    if (isLoading) {
        return <div className="min-h-screen bg-gray-50 flex items-center justify-center"><Loader2 className="w-8 h-8 animate-spin text-blue-600" /></div>;
    }

    if (!user) {
        return <AuthForm onAuthSuccess={(userData) => setUser(userData)} />;
    }

    return (
        <div className="bg-gray-50 min-h-screen">
            {notification && <Notification message={notification.message} type={notification.type} />}
            <header className="bg-white shadow-sm border-b sticky top-0 z-40">
                <div className="container mx-auto px-4 py-4 flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                        <ShieldCheck className="w-8 h-8 text-blue-600" />
                        <h1 className="text-2xl font-bold text-gray-900">Secure File Vault</h1>
                    </div>
                    <div className="flex items-center space-x-4">
                        <div className="flex items-center space-x-2 text-gray-700">
                            <User className="w-4 h-4" />
                            <span>{user.username}</span>
                        </div>
                        <button onClick={handleLogout} className="flex items-center space-x-1 text-gray-600 hover:text-gray-800 px-3 py-2 rounded-md hover:bg-gray-100"><LogOut className="w-4 h-4" /><span>Logout</span></button>
                    </div>
                </div>
            </header>
            <main className="container mx-auto px-4 py-8">
                <div className="mb-8">
                    <StorageStatsDisplay stats={storageStats} />
                </div>
                <div className="bg-white rounded-lg shadow-sm border p-6 mb-8">
                    <h2 className="text-xl font-semibold mb-4">Upload New Files</h2>
                    <UploadZone onUploadSuccess={fetchData} />
                </div>
                <div className="bg-white rounded-lg shadow-sm border p-6">
                    <h2 className="text-xl font-semibold mb-4">Manage Files</h2>
                    <SearchAndFilter onFilterChange={setActiveFilters} />
                    {error && <div className="mb-6 text-red-700 bg-red-100 p-4 rounded-lg"><AlertTriangle className="w-5 h-5 mr-2 inline" />{error}</div>}
                    <FileList files={files} onDelete={handleDeleteFile} />
                </div>
            </main>
        </div>
    );
}

export default App;