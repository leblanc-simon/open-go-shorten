const qrcodeUrl = (shortenUrl, format, download = false) =>
    `/qrcode/${encodeURIComponent(shortenUrl)}/${format}${download ? '?download=1' : ''}`;

const showQRCode = (shortenUrl) => {
    const dialog = document.createElement('dialog');
    dialog.classList.add('qrcode-dialog');
    const article = document.createElement('article');
    const main = document.createElement('main');
    const footer = document.createElement('footer');
    const close = document.createElement('button');

    const title = document.createElement('strong');
    title.textContent = `${window.location.origin}/${shortenUrl}`;

    const img = document.createElement('img');
    img.src = qrcodeUrl(shortenUrl, 'svg');
    img.alt = `QR code of ${title.textContent}`;

    const nav = document.createElement('nav');
    for (const format of ['svg', 'png']) {
        const link = document.createElement('a');
        link.href = qrcodeUrl(shortenUrl, format, true);
        link.download = `${shortenUrl}.${format}`;
        link.textContent = format === 'png' ? 'Download PNG (512x512)' : 'Download SVG';
        nav.appendChild(link);
    }

    main.append(title, img, nav);
    article.appendChild(main);

    close.textContent = 'Close';
    close.addEventListener('click', () => dialog.remove());
    dialog.addEventListener('close', () => dialog.remove());
    footer.appendChild(close);
    article.appendChild(footer);

    dialog.appendChild(article);
    document.body.appendChild(dialog);
    dialog.showModal();
};

document.addEventListener('click', event => {
    if (true === event.target.classList.contains('qrcode')) {
        event.preventDefault();
        showQRCode(event.target.dataset.qrcode);
    }
});
