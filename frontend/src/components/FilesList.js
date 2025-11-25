import React, { useState, useEffect } from 'react';

function FilesList({ onSelectFile }) {
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchFiles();
    // Poll for status updates every 2 seconds
    const interval = setInterval(fetchFiles, 2000);
    return () => clearInterval(interval);
  }, []);

  const fetchFiles = async () => {
    try {
      const response = await fetch('/api/files');
      const data = await response.json();
      setFiles(data.files || []);
      setLoading(false);
    } catch (error) {
      console.error('Error fetching files:', error);
      setLoading(false);
    }
  };

  const formatFileSize = (bytes) => {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
  };

  const formatDate = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleString();
  };

  const getStatusBadge = (status) => {
    const statusConfig = {
      processing: 'bg-yellow-100 text-yellow-800 border-yellow-200',
      completed: 'bg-green-100 text-green-800 border-green-200',
      failed: 'bg-red-100 text-red-800 border-red-200'
    };

    return (
      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${statusConfig[status] || 'bg-gray-100 text-gray-800'}`}>
        {status}
      </span>
    );
  };

  if (loading) {
    return <div className="text-center py-12 text-gray-600">Loading files...</div>;
  }

  if (files.length === 0) {
    return <div className="text-center py-12 text-gray-500">No files uploaded yet. Upload your first CSV file!</div>;
  }

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200">
      <div className="px-6 py-4 border-b border-gray-200">
        <h2 className="text-lg font-semibold text-gray-900">Uploaded CSV Files ({files.length})</h2>
      </div>
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Filename</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Size</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Records</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Uploaded</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Processing Time</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Action</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {files.map((file) => (
              <tr key={file.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{file.filename}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">{formatFileSize(file.fileSize)}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">{file.recordCount || 0}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">{formatDate(file.uploadedAt)}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">{file.processingTimeMs ? `${file.processingTimeMs}ms` : '-'}</td>
                <td className="px-6 py-4 whitespace-nowrap">{getStatusBadge(file.status)}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">
                  {file.status === 'completed' ? (
                    <button
                      className="px-4 py-2 bg-gray-900 text-white text-sm font-medium rounded-md hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-900"
                      onClick={() => onSelectFile(file)}
                    >
                      View & Search
                    </button>
                  ) : file.status === 'failed' ? (
                    <span className="text-red-600 cursor-help" title={file.errorMessage}>
                      Error
                    </span>
                  ) : (
                    <span className="text-yellow-600 font-medium">Processing...</span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export default FilesList;
