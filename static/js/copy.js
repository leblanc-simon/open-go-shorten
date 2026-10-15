const copyItem = async (button) => {
    try {
        await navigator.clipboard.writeText(button.dataset.copy);

        // Buttons with a text label display "Copied!", icon buttons a check mark
        const label = button.querySelector('[data-copy-label]');
        const use = button.querySelector('use');
        const previousLabel = label?.textContent;
        const previousIcon = use?.getAttribute('href');

        if (label) {
            label.textContent = 'Copied!';
        }
        use?.setAttribute('href', '#i-check');
        button.classList.add('is-copied');

        setTimeout(() => {
            if (label) {
                label.textContent = previousLabel;
            }
            use?.setAttribute('href', previousIcon);
            button.classList.remove('is-copied');
        }, 1200);
    } catch (error) {
        console.error(error.message);
    }
};

document.addEventListener('click', event => {
    const button = event.target.closest('[data-copy]');
    if (button && button.dataset.copy) {
        event.preventDefault();
        copyItem(button);
    }
});
