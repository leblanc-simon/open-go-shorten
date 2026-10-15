const loginContainer = document.getElementById('login-container');
const loginForm = loginContainer.querySelector('#login-form');
const passwordInput = document.getElementById('password');
const togglePasswordButton = document.getElementById('toggle-password');

const session = document.getElementById('session');
const logoutButton = document.getElementById('logout');

const shortenContainer = document.getElementById('shorten-container');
const shortenForm = shortenContainer.querySelector('#shorten-form');
const shortenedUrl = shortenContainer.querySelector('#shortened-url');

const urlsContainer = document.getElementById('urls-container');
const urlsTable = urlsContainer.querySelector('#urls-table');
const urlsTableBody = urlsTable.getElementsByTagName('tbody')[0];
const urlsCount = document.getElementById('urls-count');
const urlsEmpty = document.getElementById('urls-empty');
const searchInput = document.getElementById('search');

let token = localStorage.getItem('token');
let lastShortURL = null;

const tokenPayload = (jwt) => {
    try {
        const payload = jwt.split('.')[1].replace(/-/g, '+').replace(/_/g, '/');
        return JSON.parse(atob(payload));
    } catch (e) {
        return {};
    }
};

const tokenExpiration = (jwt) => (tokenPayload(jwt).exp || 0) * 1000;

const hideLogin = () => {
    loginContainer.classList.add('hidden');
    session.classList.remove('hidden');
    const username = tokenPayload(token).username || '';
    document.getElementById('user-name').textContent = username;
    document.getElementById('user-initial').textContent = username.charAt(0);
};
const showLogin = () => {
    loginContainer.classList.remove('hidden');
    session.classList.add('hidden');
};
const hideShorten = () => {
    shortenContainer.classList.add('hidden');
    shortenedUrl.classList.add('hidden');
};
const showShorten = () => {
    shortenContainer.classList.remove('hidden');
};
const hideUrls = () => {
    urlsContainer.classList.add('hidden');
    urlsTableBody.replaceChildren();
};
const showUrls = () => {
    urlsContainer.classList.remove('hidden');
};

const displayLogin = () => {
    showLogin();
    hideShorten();
    hideUrls();
};

const displayApp = () => {
    hideLogin();
    showShorten();
    showUrls();
    fetchURLs();
};

const logout = () => {
    localStorage.removeItem('token');
    token = null;
    displayLogin();
};

// By default, display login
displayLogin();

if (token) {
    if (tokenExpiration(token) > Date.now()) {
        displayApp()
    } else {
        localStorage.removeItem('token');
    }
}

logoutButton.addEventListener('click', logout);

togglePasswordButton.addEventListener('click', () => {
    const show = passwordInput.type === 'password';
    passwordInput.type = show ? 'text' : 'password';
    togglePasswordButton.setAttribute('aria-pressed', String(show));
    togglePasswordButton.setAttribute('aria-label', show ? 'Hide password' : 'Show password');
    togglePasswordButton.querySelector('use').setAttribute('href', show ? '#i-eye-off' : '#i-eye');
});

loginForm.addEventListener('submit', async (event) => {
    event.preventDefault();

    const username = document.getElementById('username').value;
    const password = passwordInput.value;

    const response = await fetch('/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password }),
    });

    if (response.ok) {
        const data = await response.json();
        token = data.token;
        localStorage.setItem('token', token);
        loginForm.reset();
        displayApp();
    } else {
        alert('Invalid credentials');
    }
});

shortenForm.addEventListener('submit', async (event) => {
    event.preventDefault();

    const url = document.getElementById('url').value;
    let expiration = null;
    if (document.getElementById('expiration').value) {
        expiration = new Date(document.getElementById('expiration').value + 'T23:59:59');
    }

    const response = await fetch('/shorten', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ url, expiration }),
    });

    if (response.ok) {
        const data = await response.json();
        lastShortURL = data.shortURL;
        showResult(data.shortURL, url, expiration);
        shortenForm.reset();
        fetchURLs();
    } else if (response.status === 401) {
        logout();
    } else {
        alert('Error shortening URL: ' + (await response.text()));
    }
});

searchInput.addEventListener('input', () => filterRows());

const shortURLHref = (shortURL) => `${window.location.origin}/${shortURL}`;

const qrcodeUrl = (shortURL, format, download = false) =>
    `/qrcode/${encodeURIComponent(shortURL)}/${format}${download ? '?download=1' : ''}`;

// Displays the short link: host in the text color, code in the accent color
const shortURLLabel = (shortURL) => {
    const code = document.createElement('span');
    code.textContent = shortURL;
    return [`${window.location.host}/`, code];
};

const formatDate = (date) => date.toLocaleDateString(undefined, {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
});

const svgIcon = (name) => {
    const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    svg.setAttribute('class', 'icon');
    svg.setAttribute('aria-hidden', 'true');
    const use = document.createElementNS('http://www.w3.org/2000/svg', 'use');
    use.setAttribute('href', `#i-${name}`);
    svg.appendChild(use);
    return svg;
};

const showResult = (shortURL, originalURL, expiration) => {
    const href = shortURLHref(shortURL);

    const link = document.getElementById('result-link');
    link.href = href;
    link.replaceChildren(...shortURLLabel(shortURL));

    document.getElementById('result-destination').textContent = originalURL;
    document.getElementById('result-copy').dataset.copy = href;
    document.getElementById('result-open').href = href;

    const expirationChip = document.getElementById('result-expiration');
    if (expiration) {
        expirationChip.querySelector('span').textContent = `Expires ${formatDate(expiration)}`;
        expirationChip.classList.remove('hidden');
    } else {
        expirationChip.classList.add('hidden');
    }

    const qrcode = document.getElementById('result-qrcode');
    qrcode.src = qrcodeUrl(shortURL, 'svg');
    qrcode.alt = `QR code of ${href}`;
    for (const format of ['png', 'svg']) {
        const download = document.getElementById(`result-${format}`);
        download.href = qrcodeUrl(shortURL, format, true);
        download.download = `${shortURL}.${format}`;
        download.setAttribute('aria-label', `Download the QR code as ${format.toUpperCase()}`);
    }

    shortenedUrl.classList.remove('hidden');
};

const cell = (className, label, ...children) => {
    const td = document.createElement('td');
    td.className = className;
    // Label displayed in front of the value on small screens
    if (label) {
        td.dataset.label = label;
    }
    td.append(...children);
    return td;
};

const span = (className, text) => {
    const element = document.createElement('span');
    element.className = className;
    element.textContent = text;
    return element;
};

const iconButton = (iconName, label, dataName, value, extraClass = '') => {
    const button = document.createElement('button');
    button.type = 'button';
    button.className = `btn-icon ${extraClass}`.trim();
    button.setAttribute('aria-label', label);
    button.title = label;
    button.dataset[dataName] = value;
    button.appendChild(svgIcon(iconName));
    return button;
};

const shortLinkCell = (shortURL) => {
    const container = document.createElement('div');
    container.className = 'short-link';
    const link = document.createElement('a');
    link.href = shortURLHref(shortURL);
    link.target = '_blank';
    link.rel = 'noopener noreferrer';
    link.textContent = `${window.location.host}/${shortURL}`;
    container.appendChild(link);
    if (shortURL === lastShortURL) {
        container.appendChild(span('badge', 'NEW'));
    }
    return container;
};

const destinationCell = (originalURL) => {
    const container = document.createElement('div');
    container.className = 'destination';
    container.title = originalURL;
    let host = originalURL;
    let path = '';
    try {
        const parsed = new URL(originalURL);
        host = parsed.host;
        path = parsed.pathname + parsed.search + parsed.hash;
        if (path === '/') {
            path = '';
        }
    } catch (e) {
        // keep the raw value
    }
    container.appendChild(span('destination-host', host));
    if (path) {
        container.appendChild(span('destination-path', path));
    }
    return container;
};

const expirationCell = (value) => {
    const container = document.createElement('div');
    container.className = 'expiration';
    if (!value) {
        container.appendChild(span('expiration-never', 'Never'));
        return container;
    }

    const date = new Date(value);
    container.appendChild(span('', formatDate(date)));

    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const day = new Date(date);
    day.setHours(0, 0, 0, 0);
    const days = Math.round((day - today) / 86400000);
    let hint = null;
    if (date < Date.now()) {
        hint = span('expiration-hint expired', 'Expired');
    } else if (days === 0) {
        hint = span('expiration-hint', 'Today');
    } else if (days === 1) {
        hint = span('expiration-hint', 'Tomorrow');
    } else if (days <= 7) {
        hint = span('expiration-hint', `In ${days} days`);
    }
    if (hint) {
        container.appendChild(hint);
    }
    return container;
};

const actionsCell = (shortURL) => {
    const container = document.createElement('div');
    container.className = 'row-actions';
    const label = `${window.location.host}/${shortURL}`;
    const copy = iconButton('copy', `Copy ${label}`, 'copy', shortURLHref(shortURL));
    const qrcode = iconButton('qrcode', `QR code of ${label}`, 'qrcode', shortURL);
    const remove = iconButton('trash', `Delete ${label}`, 'delete', shortURL, 'btn-icon-danger');
    container.append(copy, qrcode, remove);
    return container;
};

const updateCount = () => {
    const count = urlsTableBody.rows.length;
    urlsCount.textContent = String(count);
    urlsEmpty.classList.toggle('hidden', count > 0);
    urlsTable.classList.toggle('notloading', count === 0);
};

const filterRows = () => {
    const query = searchInput.value.trim().toLowerCase();
    for (const row of urlsTableBody.rows) {
        row.hidden = query !== '' && !row.dataset.search.includes(query);
    }
};

async function fetchURLs() {
    const response = await fetch('/urls', {
        method: 'GET',
        headers: {
            'Authorization': `Bearer ${token}`,
        },
    });

    if (response.status === 401) {
        logout();
        return;
    }
    if (!response.ok) {
        alert('Error retrieving URLs');
        return;
    }

    const urls = await response.json();
    const rows = urls.map((url) => {
        const row = document.createElement('tr');
        row.id = url.Shorten;
        row.dataset.search = `${url.Shorten} ${url.URLData.url}`.toLowerCase();

        row.append(
            cell('cell-short', null, shortLinkCell(url.Shorten)),
            cell('cell-destination', null, destinationCell(url.URLData.url)),
            cell('cell-expiration', 'Expires', expirationCell(url.URLData.expiration)),
            cell('number cell-visits', 'Visits', String(url.Stats.VisitorsCount)),
            cell('number secondary cell-unique', 'Unique', String(url.Stats.UniqueVisitorsCount)),
            cell('actions', null, actionsCell(url.Shorten)),
        );
        return row;
    });
    urlsTableBody.replaceChildren(...rows);
    updateCount();
    filterRows();
}
