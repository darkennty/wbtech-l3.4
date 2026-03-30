document.addEventListener('DOMContentLoaded', () => {
    const uploadForm = document.getElementById('uploadForm');
    const imageFile = document.getElementById('imageFile');
    const imagesList = document.getElementById('images');

    function getSavedIds() {
        return JSON.parse(localStorage.getItem('uploadedImageIds') || '[]');
    }

    function saveId(id) {
        const ids = getSavedIds();
        if (!ids.includes(id)) {
            ids.push(id);
            localStorage.setItem('uploadedImageIds', JSON.stringify(ids));
        }
    }

    function removeId(id) {
        let ids = getSavedIds();
        ids = ids.filter(savedId => savedId !== id);
        localStorage.setItem('uploadedImageIds', JSON.stringify(ids));
    }

    function restoreImages() {
        const ids = getSavedIds();
        ids.forEach(id => {
            addImageToList(id, 'restoring...');
            pollStatus(id);
        });
    }

    restoreImages();

    uploadForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData();
        formData.append('image', imageFile.files[0]);

        try {
            const res = await fetch('/upload', { method: 'POST', body: formData });
            const data = await res.json();
            if (res.ok) {
                const id = data.result.id;
                saveId(id);
                addImageToList(id, 'pending');
                pollStatus(id);
            } else {
                alert('Upload failed: ' + (data.error || 'Unknown error'));
            }
        } catch (err) {
            console.error(err);
            alert('Upload error');
        }
    });

    function addImageToList(id, status) {
        if (document.getElementById(`image-${id}`)) return;

        const li = document.createElement('li');
        li.id = `image-${id}`;
        li.innerHTML = `
            <div>
                <strong>ID:</strong> ${id}<br>
                <span class="status status-${status}">${status}</span>
            </div>
            <button class="delete" data-id="${id}">Delete</button>
        `;
        const deleteBtn = li.querySelector('.delete');
        deleteBtn.addEventListener('click', () => deleteImage(id));
        imagesList.appendChild(li);
    }

    async function pollStatus(id) {
        const interval = setInterval(async () => {
            try {
                const res = await fetch(`/image/${id}?type=status_only`);

                if (res.status === 202) {
                    const data = await res.json();
                    updateStatus(id, data.result?.status || 'processing');
                } else if (res.status === 200) {
                    clearInterval(interval);
                    updateStatus(id, 'completed');

                    const container = document.querySelector(`#image-${id} div`);
                    if (!container.querySelector('.images-wrapper')) {
                        const wrapper = document.createElement('div');
                        wrapper.className = 'images-wrapper';

                        const types = [
                            { type: 'original', title: 'Оригинал' },
                            { type: 'processed', title: 'С водяным знаком' },
                            { type: 'thumb', title: 'Миниатюра' }
                        ];

                        types.forEach(t => {
                            const imgBlock = document.createElement('div');
                            imgBlock.className = 'image-card';
                            imgBlock.innerHTML = `<h4>${t.title}</h4>`;

                            const linkElement = document.createElement('a');
                            linkElement.href = `/image/${id}?type=${t.type}`;
                            linkElement.target = '_blank';

                            const imgElement = document.createElement('img');
                            imgElement.src = `/image/${id}?type=${t.type}`;
                            imgElement.className = 'image-thumbnail';

                            linkElement.appendChild(imgElement);
                            imgBlock.appendChild(linkElement);
                            wrapper.appendChild(imgBlock);
                        });
                        container.appendChild(wrapper);
                    }
                } else {
                    console.warn(`ID ${id} not found on the server (status ${res.status}). Deleting...`);
                    clearInterval(interval);
                    removeId(id);
                    const li = document.getElementById(`image-${id}`);
                    if (li) li.remove();
                }
            } catch (err) {
                console.error("Connection error:", err);
            }
        }, 2000);
    }

    function updateStatus(id, status) {
        const span = document.querySelector(`#image-${id} .status`);
        if (span) {
            span.textContent = status;
            span.className = `status status-${status}`;
        }
    }

    async function deleteImage(id) {
        if (!confirm('Delete this image?')) return;
        try {
            const res = await fetch(`/image/${id}`, { method: 'DELETE' });
            if (res.ok) {
                const li = document.getElementById(`image-${id}`);
                if (li) li.remove();
                removeId(id);
            } else {
                alert('Delete failed');
            }
        } catch (err) {
            console.error(err);
            alert('Delete error');
        }
    }
});