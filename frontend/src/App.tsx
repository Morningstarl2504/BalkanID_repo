import React, { useState, useCallback, useMemo } from 'react';
import { useDropzone } from 'react-dropzone';
import { UploadCloud, File, Folder, BarChart, Trash2, Share2, Link as LinkIcon, Download, X, Search, Settings } from 'lucide-react';
import { BarChart as RechartsBarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';


// --- MOCK DATA (to be replaced with API calls) ---
const mockFiles = [
    { id: '1', name: 'financial_report_q1.pdf', type: 'application/pdf', size: 1024 * 500, uploader: 'You', uploadDate: '2025-09-18T10:00:00Z', isPublic: true, downloadCount: 15 },
    { id: '2', name: 'project_logo.png', type: 'image/png', size: 1024 * 120, uploader: 'You', uploadDate: '2025-09-17T15:30:00Z', isPublic: false, downloadCount: 0 },
    { id: '3', name: 'meeting_notes.docx', type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', size: 1024 * 35, uploader: 'Admin', uploadDate: '2025-09-17T11:00:00Z', isPublic: true, downloadCount: 3 },
];

const mockFolders = [
    { id: 'f1', name: 'Project Documents', files: [mockFiles[0], mockFiles[2]] },
    { id: 'f2', name: 'Marketing Assets', files: [mockFiles[1]] },
];

const mockStats = {
    totalStorageUsed: 655 * 1024,
    originalStorageUsage: 855 * 1024,
    storageSavingsBytes: 200 * 1024,
    storageSavingsPercentage: 23.39,
    quotaUsed: (655 * 1024) / (10 * 1024 * 1024) * 100,
    quotaTotal: 10 * 1024 * 1024,
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

const FileIcon = ({ type }: { type: string }) => {
    // A simple component to render a file icon based on MIME type
    if (type.startsWith('image/')) return <File className="text-blue-500" />;
    if (type === 'application/pdf') return <File className="text-red-500" />;
    return <File className="text-gray-500" />;
};


const UploadZone = () => {
    const onDrop = useCallback((acceptedFiles: File[]) => {
        // Handle file upload logic here
        console.log(acceptedFiles);
        // You would typically make an API call to upload the files here
    }, []);

    const { getRootProps, getInputProps, isDragActive } = useDropzone({ onDrop });

    return (
        <div {...getRootProps()} className={`border-2 border-dashed rounded-lg p-12 text-center cursor-pointer transition-colors ${isDragActive ? 'border-blue-500 bg-blue-50' : 'border-gray-300 hover:border-blue-400'}`}>
            <input {...getInputProps()} />
            <div className="flex flex-col items-center">
                <UploadCloud className="w-12 h-12 text-gray-400 mb-4" />
                {isDragActive ?
                    <p className="text-lg text-blue-600">Drop the files here ...</p> :
                    <p className="text-lg text-gray-600">Drag & drop some files here, or click to select files</p>
                }
                <p className="text-sm text-gray-500 mt-2">Single or multiple file upload supported</p>
            </div>
        </div>
    );
};


const FileList = ({ files, folders }: { files: any[], folders: any[] }) => {
    const [selectedItem, setSelectedItem] = useState<any>(null);

    const openShareModal = (item: any) => {
        setSelectedItem(item);
    };

    const closeShareModal = () => {
        setSelectedItem(null);
    };

    return (
        <div className="mt-8">
            <h3 className="text-xl font-semibold mb-4 text-gray-800">Your Files</h3>
             {/* Search and Filter Bar */}
             <div className="flex flex-wrap gap-4 mb-6 p-4 bg-gray-50 rounded-lg border">
                <div className="relative flex-grow">
                    <input type="text" placeholder="Search by filename..." className="w-full p-2 pl-10 border rounded-md focus:ring-2 focus:ring-blue-500 outline-none" />
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                </div>
                <select className="p-2 border rounded-md bg-white">
                    <option>Filter by type</option>
                    <option>image/png</option>
                    <option>application/pdf</option>
                    <option>document</option>
                </select>
                <input type="date" className="p-2 border rounded-md bg-white" />
                <button className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition">Apply Filters</button>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                {/* Render Folders */}
                {folders.map(folder => (
                    <div key={folder.id} className="bg-white p-4 rounded-lg shadow-sm border hover:shadow-md transition-shadow flex items-center space-x-4">
                        <Folder className="w-10 h-10 text-yellow-500" />
                        <div className="flex-grow">
                            <p className="font-semibold text-gray-800">{folder.name}</p>
                            <p className="text-sm text-gray-500">{folder.files.length} files</p>
                        </div>
                        <button onClick={() => openShareModal(folder)} className="p-2 text-gray-500 hover:text-blue-600"><Share2 size={18} /></button>
                    </div>
                ))}

                {/* Render Files */}
                {files.map(file => (
                    <div key={file.id} className="bg-white p-4 rounded-lg shadow-sm border hover:shadow-md transition-shadow">
                        <div className="flex items-start space-x-4">
                            <FileIcon type={file.type} />
                            <div className="flex-grow overflow-hidden">
                                <p className="font-semibold text-gray-800 truncate" title={file.name}>{file.name}</p>
                                <p className="text-sm text-gray-500">{formatBytes(file.size)}</p>
                            </div>
                        </div>
                        <div className="text-xs text-gray-400 mt-2">
                            <p>Uploader: {file.uploader}</p>
                            <p>Date: {new Date(file.uploadDate).toLocaleDateString()}</p>
                        </div>
                        <div className="flex justify-between items-center mt-4 pt-3 border-t">
                             <div className="flex items-center space-x-2 text-sm text-gray-500">
                                {file.isPublic && (
                                    <>
                                        <Download size={16} />
                                        <span>{file.downloadCount}</span>
                                    </>
                                )}
                            </div>
                            <div className="flex space-x-2">
                                <button onClick={() => openShareModal(file)} className="p-2 text-gray-500 hover:text-blue-600"><Share2 size={18} /></button>
                                <button className="p-2 text-gray-500 hover:text-red-600"><Trash2 size={18} /></button>
                            </div>
                        </div>
                    </div>
                ))}
            </div>
            {selectedItem && <ShareModal item={selectedItem} onClose={closeShareModal} />}
        </div>
    );
};

const ShareModal = ({ item, onClose }: { item: any; onClose: () => void }) => {
    const publicLink = `https://filevault.example.com/share/${item.id}`;
    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg shadow-xl p-8 w-full max-w-md m-4">
                <div className="flex justify-between items-center mb-4">
                    <h3 className="text-xl font-bold">Share "{item.name}"</h3>
                    <button onClick={onClose} className="text-gray-500 hover:text-gray-800">
                        <X size={24} />
                    </button>
                </div>
                
                {/* Public Sharing */}
                <div className="mb-6">
                    <h4 className="font-semibold mb-2">Public Link</h4>
                    <div className="flex items-center space-x-2">
                        <input type="text" readOnly value={publicLink} className="w-full p-2 border rounded-md bg-gray-100" />
                        <button 
                            onClick={() => navigator.clipboard.writeText(publicLink)}
                            className="bg-blue-500 text-white p-2 rounded-md hover:bg-blue-600"
                        >
                            <LinkIcon size={20} />
                        </button>
                    </div>
                </div>

                {/* Share with specific users (Bonus) */}
                <div>
                    <h4 className="font-semibold mb-2">Share with specific users</h4>
                    <div className="flex items-center space-x-2">
                        <input type="email" placeholder="Enter user email" className="w-full p-2 border rounded-md" />
                        <button className="bg-green-500 text-white px-4 py-2 rounded-md hover:bg-green-600">Share</button>
                    </div>
                </div>
                 <div className="mt-4 text-sm text-gray-600">
                    <p>Current collaborators:</p>
                    <ul className="list-disc list-inside">
                        <li>user1@example.com (View)</li>
                        <li>user2@example.com (Edit)</li>
                    </ul>
                </div>
            </div>
        </div>
    );
};

const StorageStats = ({ stats }: { stats: typeof mockStats }) => {
    const data = [
        { name: 'Used', value: stats.totalStorageUsed },
        { name: 'Saved', value: stats.storageSavingsBytes },
    ];

    return (
        <div className="mt-8 p-6 bg-white rounded-lg shadow-sm border">
            <h3 className="text-xl font-semibold mb-4 text-gray-800">Storage Statistics</h3>
            <div className="grid md:grid-cols-2 gap-6">
                <div>
                    <p className="text-gray-600">Total Storage Used: <span className="font-bold text-blue-600">{formatBytes(stats.totalStorageUsed)}</span></p>
                    <p className="text-gray-600">Original Usage: <span className="font-bold">{formatBytes(stats.originalStorageUsage)}</span></p>
                    <p className="text-gray-600">Deduplication Savings: <span className="font-bold text-green-600">{formatBytes(stats.storageSavingsBytes)} ({stats.storageSavingsPercentage}%)</span></p>
                    <div className="w-full bg-gray-200 rounded-full h-2.5 my-4">
                        <div className="bg-blue-600 h-2.5 rounded-full" style={{ width: `${stats.quotaUsed}%` }}></div>
                    </div>
                     <p className="text-sm text-gray-500 text-right">{formatBytes(stats.totalStorageUsed)} of {formatBytes(stats.quotaTotal)} used</p>
                </div>
                <div style={{ width: '100%', height: 200 }}>
                    <ResponsiveContainer>
                        <RechartsBarChart data={data}>
                            <CartesianGrid strokeDasharray="3 3" />
                            <XAxis dataKey="name" />
                            <YAxis tickFormatter={(tick) => formatBytes(tick)} />
                            <Tooltip formatter={(value: number) => formatBytes(value)} />
                            <Legend />
                            <Bar dataKey="value" fill="#3B82F6" name="Storage" />
                        </RechartsBarChart>
                    </ResponsiveContainer>
                </div>
            </div>
        </div>
    );
};

// --- ADMIN PANEL COMPONENTS ---
const AdminPanel = () => {
     const allFiles = useMemo(() => [
        ...mockFiles,
        { id: '4', name: 'admin_only_doc.pdf', type: 'application/pdf', size: 1024 * 720, uploader: 'Admin', uploadDate: '2025-09-18T12:00:00Z', isPublic: false, downloadCount: 1 },
        { id: '5', name: 'company_policy.docx', type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', size: 1024 * 50, uploader: 'Admin', uploadDate: '2025-09-16T09:00:00Z', isPublic: true, downloadCount: 42 },
     ], []);

    return (
        <div className="mt-8 p-6 bg-red-50 rounded-lg border border-red-200">
             <h3 className="text-xl font-semibold mb-4 text-red-800">Admin Panel</h3>
             <div className="overflow-x-auto">
                 <table className="min-w-full bg-white rounded-lg">
                    <thead className="bg-gray-100">
                        <tr>
                            <th className="text-left p-3">Filename</th>
                            <th className="text-left p-3">Uploader</th>
                            <th className="text-left p-3">Size</th>
                            <th className="text-left p-3">Downloads</th>
                            <th className="text-left p-3">Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        {allFiles.map(file => (
                            <tr key={file.id} className="border-b">
                                <td className="p-3 font-medium">{file.name}</td>
                                <td className="p-3">{file.uploader}</td>
                                <td className="p-3">{formatBytes(file.size)}</td>
                                <td className="p-3">{file.downloadCount}</td>
                                <td className="p-3">
                                    <button className="text-blue-600 hover:underline">Share</button>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                 </table>
             </div>
        </div>
    );
};


// --- MAIN APP COMPONENT ---

function App() {
    const [isAdmin, setIsAdmin] = useState(false);

    return (
        <div className="bg-gray-50 min-h-screen font-sans">
            <header className="bg-white shadow-sm">
                <div className="container mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="flex justify-between items-center py-4">
                        <h1 className="text-2xl font-bold text-gray-900">
                           <Folder className="inline-block w-8 h-8 mr-2 text-blue-600" /> Secure File Vault
                        </h1>
                        <div className="flex items-center space-x-4">
                             <label htmlFor="admin-toggle" className="flex items-center cursor-pointer">
                                <span className="mr-3 text-sm font-medium text-gray-900">Admin Mode</span>
                                <div className="relative">
                                    <input type="checkbox" id="admin-toggle" className="sr-only" checked={isAdmin} onChange={() => setIsAdmin(!isAdmin)} />
                                    <div className="block bg-gray-600 w-14 h-8 rounded-full"></div>
                                    <div className={`dot absolute left-1 top-1 bg-white w-6 h-6 rounded-full transition ${isAdmin ? 'transform translate-x-full bg-blue-600' : ''}`}></div>
                                </div>
                            </label>
                            <button className="p-2 rounded-full hover:bg-gray-100">
                                <Settings />
                            </button>
                        </div>
                    </div>
                </div>
            </header>

            <main className="container mx-auto px-4 sm:px-6 lg:px-8 py-8">
                <div className="max-w-7xl mx-auto">
                    {/* File Upload Section */}
                    <div className="bg-white p-6 rounded-lg shadow-sm border mb-8">
                         <h2 className="text-xl font-semibold mb-4 text-gray-800">Upload New Files</h2>
                        <UploadZone />
                    </div>

                    {/* File List and Management */}
                    <FileList files={mockFiles} folders={mockFolders} />

                    {/* Storage Statistics */}
                    <StorageStats stats={mockStats} />

                    {/* Admin Panel (Conditional) */}
                    {isAdmin && <AdminPanel />}

                </div>
            </main>
        </div>
    );
}

export default App;
