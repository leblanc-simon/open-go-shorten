const showQRCode = (shortURL) => {
    const href = shortURLHref(shortURL);

    const dialog = document.createElement('dialog');
    dialog.classList.add('qrcode-dialog');
    dialog.setAttribute('aria-labelledby', 'qrcode-title');
    const article = document.createElement('article');

    const header = document.createElement('header');
    header.className = 'dialog-header';
    const heading = document.createElement('div');
    const title = document.createElement('h2');
    title.id = 'qrcode-title';
    title.textContent = 'QR code';
    const subtitle = document.createElement('p');
    subtitle.className = 'dialog-subtitle';
    subtitle.textContent = 'Scan to open the short link';
    heading.append(title, subtitle);

    const close = document.createElement('button');
    close.type = 'button';
    close.className = 'btn-icon';
    close.setAttribute('aria-label', 'Close');
    close.appendChild(svgIcon('close'));
    close.addEventListener('click', () => dialog.close());
    header.append(heading, close);

    const frame = document.createElement('div');
    frame.className = 'qrcode-frame';
    const img = document.createElement('img');
    img.src = qrcodeUrl(shortURL, 'svg');
    img.alt = `QR code of ${href}`;
    frame.appendChild(img);

    const link = document.createElement('p');
    link.className = 'qrcode-link';
    link.append(...shortURLLabel(shortURL));

    const actions = document.createElement('div');
    actions.className = 'qrcode-actions';
    for (const format of ['png', 'svg']) {
        const download = document.createElement('a');
        download.className = format === 'png' ? 'btn btn-primary' : 'btn btn-ghost';
        download.href = qrcodeUrl(shortURL, format, true);
        download.download = `${shortURL}.${format}`;
        download.append(svgIcon('download'), format === 'png' ? 'Download PNG' : 'Download SVG');
        if (format === 'png') {
            download.title = '512x512 pixels';
        }
        actions.appendChild(download);
    }

    article.append(header, frame, link, actions);
    dialog.appendChild(article);
    dialog.addEventListener('close', () => dialog.remove());
    dialog.addEventListener('click', (event) => {
        // Click on the backdrop
        if (event.target === dialog) {
            dialog.close();
        }
    });
    document.body.appendChild(dialog);
    dialog.showModal();
};

document.addEventListener('click', event => {
    const button = event.target.closest('[data-qrcode]');
    if (button) {
        event.preventDefault();
        showQRCode(button.dataset.qrcode);
    }
});
