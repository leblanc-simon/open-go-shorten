const loginContainer = document.getElementById('login-container');
const loginForm = loginContainer.querySelector('#login-form');
const logoutButton = document.getElementById('logout');

const shortenContainer = document.getElementById('shorten-container');
const shortenForm = shortenContainer.querySelector('#shorten-form');
const shortenedUrl = shortenContainer.querySelector('#shortened-url');

const urlsContainer = document.getElementById('urls-container');
const urlsTable = urlsContainer.querySelector('#urls-table');
const urlsTableBody = urlsTable.getElementsByTagName('tbody')[0];

let token = localStorage.getItem('token');

const hideLogin = () => {
    loginContainer.classList.add('hidden');
    logoutButton.classList.remove('hidden');
};
const showLogin = () => {
    loginContainer.classList.remove('hidden');
    logoutButton.classList.add('hidden');
};
const hideShorten = () => {
    shortenContainer.classList.add('hidden');
};
const showShorten = () => {
    shortenContainer.classList.remove('hidden');
};
const hideUrls = () => {
    urlsContainer.classList.add('hidden');
    urlsTable.querySelector('tbody').innerHTML = '';
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
    displayLogin();
};

// By default, display login
displayLogin();

const tokenExpiration = (jwt) => {
    try {
        const payload = jwt.split('.')[1].replace(/-/g, '+').replace(/_/g, '/');
        return JSON.parse(atob(payload)).exp * 1000;
    } catch (e) {
        return 0;
    }
};

if (token) {
    if (tokenExpiration(token) > Date.now()) {
        displayApp()
    } else {
        localStorage.removeItem('token');
    }
}

logoutButton.addEventListener('click', logout);

loginForm.addEventListener('submit', async (event) => {
    event.preventDefault();

    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

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
        const qrcode = document.createElement('i');
        qrcode.className = 'qrcode';
        qrcode.dataset.qrcode = data.shortURL;
        qrcode.title = 'QR code';
        const qrcodeContainer = document.createElement('span');
        qrcodeContainer.appendChild(qrcode);
        shortenedUrl.replaceChildren('Shortened URL:', shortLink(data.shortURL), qrcodeContainer);
        shortenForm.reset();
        fetchURLs();
    } else if (response.status === 401) {
        logout();
    } else {
        alert('Error shortening URL: ' + (await response.text()));
    }
});

const shortLink = (shortURL) => {
    const link = document.createElement('a');
    link.href = `${window.location.origin}/${shortURL}`;
    link.target = '_blank';
    link.rel = 'noopener noreferrer';
    const copy = document.createElement('span');
    copy.className = 'copy';
    copy.dataset.copyText = 'Copy URL';
    link.append(copy, link.href);
    return link;
};

const cell = (...children) => {
    const td = document.createElement('td');
    td.append(...children);
    return td;
};

const icon = (className, dataName, shortURL, title) => {
    const i = document.createElement('i');
    i.className = className;
    i.dataset[dataName] = shortURL;
    i.title = title;
    return i;
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

        const originalURL = document.createElement('div');
        originalURL.className = 'overflow';
        originalURL.textContent = url.URLData.url;

        const expiration = url.URLData.expiration
            ? new Date(url.URLData.expiration).toLocaleString()
            : 'No expiration';

        row.append(
            cell(shortLink(url.Shorten)),
            cell(originalURL),
            cell(expiration),
            cell(String(url.Stats.VisitorsCount)),
            cell(String(url.Stats.UniqueVisitorsCount)),
            cell(
                icon('qrcode', 'qrcode', url.Shorten, 'QR code'),
                icon('delete', 'delete', url.Shorten, 'Delete'),
            ),
        );
        return row;
    });
    urlsTableBody.replaceChildren(...rows);
    urlsTable.classList.remove('notloading');
}
