const confirmDelete = (shortenUrl) => {
    const dialog = document.createElement('dialog');
    const article = document.createElement('article');
    const main = document.createElement('main');
    const footer = document.createElement('footer');
    const cancel = document.createElement('button');
    const ok = document.createElement('button');

    const link = document.createElement('strong');
    link.textContent = `${window.location.origin}/${shortenUrl}`;
    main.append('Could you delete the short link : ', link, ' ?');
    article.appendChild(main);

    cancel.addEventListener('click', closeDelete, true);
    cancel.innerText = 'Cancel';
    ok.addEventListener('click', async (event) => {
        event.preventDefault();
        try {
            const response = await fetch('/' + encodeURIComponent(shortenUrl), {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${token}`,
                },
            });
            if (response.status === 401) {
                closeDelete(dialog);
                logout();
                return;
            }
            if (!response.ok) {
                alert('Error deleting URL');
                return;
            }

            const jsonResponse = await response.json();
            document.getElementById(jsonResponse.shortURL)?.remove();
        } catch (e) {
            console.error(e);
        }
        closeDelete(dialog);
    });
    ok.innerText = 'Delete';
    ok.dataset.remove = true;

    footer.appendChild(cancel);
    footer.appendChild(ok);

    article.appendChild(footer);

    dialog.appendChild(article);

    document.body.appendChild(dialog);
    dialog.showModal();
};

const closeDelete = (event) => {
    let dialog;
    if (event.type === 'click') {
        dialog = event.target.closest('dialog');
    } else if (event.nodeName === 'DIALOG') {
        dialog = event;
    }

    dialog?.remove();
}


const deleteItem = async (event) => {
    try {
        event.preventDefault();
        const shortenUrl = event.target.dataset.delete;
        confirmDelete(shortenUrl);
    } catch (error) {
        console.error(error.message);
    }
};

document.addEventListener('click', event => {
    if (true === event.target.classList.contains('delete')) {
        deleteItem(event);
    }
});
