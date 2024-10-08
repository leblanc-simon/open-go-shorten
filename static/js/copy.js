const copyItem = async (event) => {
    try {
        event.preventDefault();
        const currentLabel = event.target.dataset.copyText;
        let copyContent = event.target?.nextElementSibling?.href;
        if (!copyContent) {
            copyContent = event.target?.parentElement?.href;
        }

        await navigator.clipboard.writeText(copyContent);
        
        event.target.dataset.copyText = 'Copied !';
        setTimeout(() => {
            event.target.dataset.copyText = currentLabel;
        }, 1000);
    } catch (error) {
        console.error(error.message);
    }
};

document.addEventListener('click', event => {
    if (true === event.target.classList.contains('copy')) {
        copyItem(event);
    }
});