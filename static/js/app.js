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

if (token) {
    const payload = JSON.parse(atob(token.split('.')[1]));
    const expirationTime = payload.exp * 1000;
    if (expirationTime > Date.now()) {
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
        shortenedUrl.innerHTML = `
        Shortened URL:
        <a href="${window.location.origin}/${data.shortURL}" target="_blank" rel="noopener noreferer">
            <span class="copy" data-copy-text="Copy URL"></span>
            ${window.location.origin}/${data.shortURL}
        </a>
        `;
        fetchURLs();
    } else {
        alert('Error shortening URL');
    }
});

async function fetchURLs() {
    const response = await fetch('/urls', {
        method: 'GET',
        headers: {
            'Authorization': `Bearer ${token}`,
        },
    });

    if (response.ok) {
        const urls = await response.json();
        urlsTableBody.innerHTML = '';
        for (const url of urls) {
            const row = document.createElement('tr');
            row.id = url.Shorten;
            let expiration;
            if ("1970-01-01T01:00:00.000000001+01:00" === url.URLData.expiration) {
                expiration = 'No expiration';
            } else {
                expiration = new Date(url.URLData.expiration).toLocaleString();
            }
            row.innerHTML = `
    <td>
        <span class="copy" data-copy-text="Copy URL"></span>
        <a href="${window.location.origin}/${url.Shorten}" target="_blank" rel="noopener noreferer">${window.location.origin}/${url.Shorten}</a>
    </td>
    <td>${url.URLData.url}</td>
    <td>${expiration}</td>
    <td>${url.Stats.VisitorsCount}</td>
    <td>${url.Stats.UniqueVisitorsCount}</td>
    <td><i class="delete" data-delete="${url.Shorten}"></i></td>
    `;
            urlsTableBody.appendChild(row);
        }
        urlsTable.classList.remove('notloading');
    } else {
        alert('Error retrieving URLs');
    }
}
