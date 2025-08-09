(function() {
    const themeToggle = document.getElementById('theme-toggle');
    const prefersDarkScheme = window.matchMedia('(prefers-color-scheme: dark)');
    
    function initTheme() {
        const savedTheme = localStorage.getItem('theme');
        
        if (savedTheme) {
            document.body.setAttribute('data-theme', savedTheme);
        } else if (prefersDarkScheme.matches) {
            document.body.setAttribute('data-theme', 'dark');
        } else {
            document.body.setAttribute('data-theme', 'light');
        }
    }
    
    function toggleTheme() {
        const currentTheme = document.body.getAttribute('data-theme');
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
        
        document.body.setAttribute('data-theme', newTheme);
        localStorage.setItem('theme', newTheme);
    }
    
    initTheme();
    
    themeToggle.addEventListener('click', toggleTheme);
    
    prefersDarkScheme.addEventListener('change', (e) => {
        if (!localStorage.getItem('theme')) {
            document.body.setAttribute('data-theme', e.matches ? 'dark' : 'light');
        }
    });
    
    const editor = document.getElementById('editor');
    const status = document.getElementById('status');
    
    let saveTimeout = null;
    let isLoading = false;
    
    const AUTOSAVE_DELAY = 500;
    
    function updateStatus(text, className = '') {
        status.textContent = text;
        status.className = 'status';
        if (className) {
            status.classList.add(className);
        }
    }
    
    async function loadContent() {
        try {
            updateStatus('Loading...', 'saving');
            const response = await fetch('/api/content');
            if (!response.ok) {
                throw new Error('Failed to load content');
            }
            const data = await response.json();
            editor.value = data.content || '';
            
            if (data.attachments) {
                displayFiles(data.attachments);
            }
            
            if (data.expiresAt) {
                displayExpiration(data.expiresAt);
            }
            
            updateStatus('Ready', 'saved');
        } catch (error) {
            console.error('Error loading content:', error);
            updateStatus('Error loading', 'error');
            setTimeout(() => updateStatus('Ready', 'saved'), 3000);
        }
    }
    
    async function saveContent() {
        if (isLoading) return;
        
        try {
            isLoading = true;
            updateStatus('Saving...', 'saving');
            
            const response = await fetch('/api/content', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    content: editor.value
                })
            });
            
            if (!response.ok) {
                throw new Error('Failed to save content');
            }
            
            updateStatus('Saved', 'saved');
            setTimeout(() => updateStatus('Ready', 'saved'), 2000);
        } catch (error) {
            console.error('Error saving content:', error);
            updateStatus('Error saving', 'error');
            setTimeout(() => updateStatus('Ready', 'saved'), 3000);
        } finally {
            isLoading = false;
        }
    }
    
    function scheduleSave() {
        if (saveTimeout) {
            clearTimeout(saveTimeout);
        }
        
        updateStatus('Typing...', 'saving');
        
        saveTimeout = setTimeout(() => {
            saveContent();
        }, AUTOSAVE_DELAY);
    }
    
    editor.addEventListener('input', scheduleSave);
    
    editor.addEventListener('paste', (e) => {
        setTimeout(() => scheduleSave(), 0);
    });
    
    window.addEventListener('beforeunload', (e) => {
        if (saveTimeout) {
            clearTimeout(saveTimeout);
            saveContent();
        }
    });
    
    document.addEventListener('visibilitychange', () => {
        if (document.hidden && saveTimeout) {
            clearTimeout(saveTimeout);
            saveContent();
        } else if (!document.hidden) {
            loadContent();
        }
    });
    
    const fileInput = document.getElementById('file-input');
    const filesList = document.getElementById('files-list');
    
    function formatFileSize(bytes) {
        if (bytes < 1024) return bytes + ' B';
        if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
        return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    }
    
    function formatExpiration(expiresAt) {
        const expDate = new Date(expiresAt);
        const now = new Date();
        const diff = expDate - now;
        
        if (diff <= 0) {
            return 'Expired';
        }
        
        const hours = Math.floor(diff / (1000 * 60 * 60));
        const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
        
        if (hours >= 24) {
            const days = Math.floor(hours / 24);
            return `Expires in ${days} day${days === 1 ? '' : 's'}`;
        } else if (hours > 0) {
            return `Expires in ${hours}h ${minutes}m`;
        } else {
            return `Expires in ${minutes} minute${minutes === 1 ? '' : 's'}`;
        }
    }
    
    function displayExpiration(expiresAt) {
        const expirationDiv = document.getElementById('expiration-display');
        if (expirationDiv) {
            const expirationText = formatExpiration(expiresAt);
            expirationDiv.textContent = expirationText;
            
            setTimeout(() => displayExpiration(expiresAt), 60000);
        }
    }
    
    function displayFiles(files) {
        filesList.innerHTML = '';
        
        if (!files || files.length === 0) {
            filesList.innerHTML = '<p style="text-align: center; color: #999; padding: 20px;">No files uploaded</p>';
            return;
        }
        
        files.forEach(file => {
            const fileItem = document.createElement('div');
            fileItem.className = 'file-item';
            
            const expirationText = file.expiresAt ? formatExpiration(file.expiresAt) : '';
            fileItem.innerHTML = `
                <div class="file-info">
                    <div class="file-name" title="${file.fileName}">${file.fileName}</div>
                    <div class="file-meta">
                        <span class="file-size">${formatFileSize(file.size)}</span>
                        <span class="file-expiration">${expirationText}</span>
                    </div>
                </div>
                <div class="file-actions">
                    <button class="file-action-btn download-btn" data-id="${file.id}" data-name="${file.fileName}" aria-label="Download file" title="Download">
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                            <polyline points="7 10 12 15 17 10"></polyline>
                            <line x1="12" y1="15" x2="12" y2="3"></line>
                        </svg>
                    </button>
                    <button class="file-action-btn delete-btn" data-id="${file.id}" aria-label="Delete file" title="Delete">
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <polyline points="3 6 5 6 21 6"></polyline>
                            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                            <line x1="10" y1="11" x2="10" y2="17"></line>
                            <line x1="14" y1="11" x2="14" y2="17"></line>
                        </svg>
                    </button>
                </div>
            `;
            
            filesList.appendChild(fileItem);
        });
        
        document.querySelectorAll('.download-btn').forEach(btn => {
            btn.addEventListener('click', handleDownload);
        });
        
        document.querySelectorAll('.delete-btn').forEach(btn => {
            btn.addEventListener('click', handleDelete);
        });
    }
    
    async function handleUpload(event) {
        const files = event.target.files;
        if (!files || files.length === 0) return;
        
        for (const file of files) {
            try {
                updateStatus('Uploading...', 'saving');
                
                const formData = new FormData();
                formData.append('file', file);
                
                const response = await fetch('/api/files/upload', {
                    method: 'POST',
                    body: formData
                });
                
                const data = await response.json();
                
                if (!response.ok || !data.success) {
                    const errorMsg = data.error || 'Failed to upload file';
                    throw new Error(errorMsg);
                }
                
                updateStatus('File uploaded', 'saved');
                loadFiles();
            } catch (error) {
                console.error('Error uploading file:', error);
                const errorMessage = error.message || 'Upload failed';
                updateStatus(`Error: ${errorMessage}`, 'error');
            }
        }
        
        event.target.value = '';
        setTimeout(() => updateStatus('Ready', 'saved'), 2000);
    }
    
    async function handleDownload(event) {
        const fileId = event.target.dataset.id;
        const fileName = event.target.dataset.name;
        
        try {
            const response = await fetch(`/api/files/${fileId}`);
            if (!response.ok) {
                throw new Error('Failed to download file');
            }
            
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = fileName;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
        } catch (error) {
            console.error('Error downloading file:', error);
            updateStatus('Download failed', 'error');
            setTimeout(() => updateStatus('Ready', 'saved'), 2000);
        }
    }
    
    async function handleDelete(event) {
        const fileId = event.target.dataset.id;
        
        if (!confirm('Are you sure you want to delete this file?')) {
            return;
        }
        
        try {
            updateStatus('Deleting...', 'saving');
            
            const response = await fetch(`/api/files/${fileId}`, {
                method: 'DELETE'
            });
            
            if (!response.ok) {
                throw new Error('Failed to delete file');
            }
            
            updateStatus('File deleted', 'saved');
            loadFiles();
            setTimeout(() => updateStatus('Ready', 'saved'), 2000);
        } catch (error) {
            console.error('Error deleting file:', error);
            updateStatus('Delete failed', 'error');
            setTimeout(() => updateStatus('Ready', 'saved'), 2000);
        }
    }
    
    async function loadFiles() {
        try {
            const response = await fetch('/api/files');
            if (!response.ok) {
                throw new Error('Failed to load files');
            }
            
            const files = await response.json();
            displayFiles(files);
        } catch (error) {
            console.error('Error loading files:', error);
        }
    }
    
    fileInput.addEventListener('change', handleUpload);
    
    loadContent();
})();