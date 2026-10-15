const confirmDelete = (shortURL) => {
    const dialog = document.createElement('dialog');
    dialog.setAttribute('role', 'alertdialog');
    dialog.setAttribute('aria-labelledby', 'delete-title');
    dialog.setAttribute('aria-describedby', 'delete-description');
    const article = document.createElement('article');

    const icon = document.createElement('div');
    icon.className = 'dialog-icon dialog-icon-danger';
    icon.appendChild(svgIcon('trash'));

    const text = document.createElement('div');
    text.className = 'dialog-text';
    const title = document.createElement('h2');
    title.id = 'delete-title';
    title.textContent = 'Delete this short link?';
    const description = document.createElement('p');
    description.id = 'delete-description';
    description.textContent = "The link and its QR code will stop working immediately. This can't be undone.";
    text.append(title, description);

    const code = document.createElement('div');
    code.className = 'dialog-code';
    code.textContent = `${window.location.host}/${shortURL}`;

    const footer = document.createElement('footer');
    const cancel = document.createElement('button');
    cancel.type = 'button';
    cancel.className = 'btn btn-ghost';
    cancel.textContent = 'Cancel';
    cancel.addEventListener('click', () => dialog.close());

    const ok = document.createElement('button');
    ok.type = 'button';
    ok.className = 'btn btn-danger';
    ok.append(svgIcon('trash'), 'Delete link');
    ok.addEventListener('click', async (event) => {
        event.preventDefault();
        ok.disabled = true;
        try {
            const response = await fetch('/' + encodeURIComponent(shortURL), {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${token}`,
                },
            });
            if (response.status === 401) {
                dialog.close();
                logout();
                return;
            }
            if (!response.ok) {
                ok.disabled = false;
                alert('Error deleting URL');
                return;
            }

            const jsonResponse = await response.json();
            document.getElementById(jsonResponse.shortURL)?.remove();
            updateCount();
        } catch (e) {
            console.error(e);
        }
        dialog.close();
    });

    footer.append(cancel, ok);
    article.append(icon, text, code, footer);
    dialog.appendChild(article);
    dialog.addEventListener('close', () => dialog.remove());

    document.body.appendChild(dialog);
    dialog.showModal();
    cancel.focus();
};

document.addEventListener('click', event => {
    const button = event.target.closest('[data-delete]');
    if (button) {
        event.preventDefault();
        confirmDelete(button.dataset.delete);
    }
});
