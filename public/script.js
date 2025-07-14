const API_URL = ''; // Так как фронт и бэк на одном домене, можно оставить пустым

// --- Вспомогательные функции ---
async function apiRequest(endpoint, method, body = null) {
    const headers = {
        'Content-Type': 'application/json',
    };
    const token = localStorage.getItem('token');
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    const config = {
        method: method,
        headers: headers,
    };

    if (body) {
        config.body = JSON.stringify(body);
    }

    try {
        const response = await fetch(API_URL + endpoint, config);
        if (response.status === 401) {
            logout();
            return;
        }
        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Что-то пошло не так');
        }
        if (response.status === 204 || response.headers.get('Content-Length') === '0') {
             return null;
        }
        return await response.json();
    } catch (error) {
        alert(error.message);
        throw error;
    }
}

function logout() {
    localStorage.removeItem('token');
    window.location.href = '/';
}


// --- Логика для страницы входа (index.html) ---
const loginForm = document.getElementById('login-form');
const registerForm = document.getElementById('register-form');

function toggleForms() {
    loginForm.classList.toggle('hidden');
    registerForm.classList.toggle('hidden');
}

async function login(event) {
    event.preventDefault();
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;
    try {
        const data = await apiRequest('/login', 'POST', { email, password });
        localStorage.setItem('token', data.token);
        window.location.href = '/cards.html';
    } catch (error) {
        // Ошибка уже показана в apiRequest
    }
}

async function register(event) {
    event.preventDefault();
    const email = document.getElementById('register-email').value;
    const password = document.getElementById('register-password').value;
    try {
        await apiRequest('/register', 'POST', { email, password });
        alert('Регистрация прошла успешно! Теперь вы можете войти.');
        toggleForms();
        document.getElementById('login-email').value = email;
        document.getElementById('login-password').focus();
    } catch (error) {
       // Ошибка уже показана в apiRequest
    }
}

// --- Логика для страницы карточек (cards.html) ---
if (window.location.pathname.endsWith('cards.html')) {
    const modal = document.getElementById('card-modal');
    const addCardBtn = document.getElementById('add-card-btn');
    const closeBtn = document.querySelector('.close-btn');
    const cardForm = document.getElementById('card-form');
    const modalTitle = document.getElementById('modal-title');
    const cardIdInput = document.getElementById('card-id');
    
    addCardBtn.onclick = () => {
        modalTitle.innerText = "Новая карточка";
        cardForm.reset();
        cardIdInput.value = '';
        modal.classList.remove('hidden');
    }
    
    closeBtn.onclick = () => modal.classList.add('hidden');
    window.onclick = (event) => {
        if (event.target == modal) {
            modal.classList.add('hidden');
        }
    }
    
    cardForm.onsubmit = async (event) => {
        event.preventDefault();
        const id = cardIdInput.value;
        const cardData = {
            word: document.getElementById('card-word').value,
            meaning: document.getElementById('card-meaning').value,
            example: document.getElementById('card-example').value,
        };

        try {
            if (id) {
                await apiRequest(`/cards/${id}`, 'PUT', cardData);
            } else {
                await apiRequest('/cards', 'POST', cardData);
            }
            modal.classList.add('hidden');
            loadCards();
        } catch (error) {
             // Ошибка уже показана в apiRequest
        }
    };
}


async function loadCards() {
    try {
        const cards = await apiRequest('/cards', 'GET');
        const container = document.getElementById('cards-container');
        container.innerHTML = '';
        if (cards && cards.length > 0) {
            cards.forEach(card => {
                const cardElement = document.createElement('div');
                cardElement.className = 'card';
                cardElement.innerHTML = `
                    <div>
                        <h3>${card.word}</h3>
                        <p>${card.meaning}</p>
                        ${card.example ? `<em>${card.example}</em>` : ''}
                    </div>
                    <div class="card-actions">
                        <button class="edit-btn" onclick="editCard(${card.id})">Изменить</button>
                        <button class="delete-btn" onclick="deleteCard(${card.id})">Удалить</button>
                    </div>
                `;
                container.appendChild(cardElement);
            });
        } else {
            container.innerHTML = '<p>У вас пока нет карточек. Добавьте первую!</p>';
        }
    } catch (error) {
       // Ошибка уже показана в apiRequest
    }
}

async function editCard(id) {
    try {
        const card = await apiRequest(`/cards/${id}`, 'GET');
        document.getElementById('modal-title').innerText = "Редактировать карточку";
        document.getElementById('card-id').value = card.id;
        document.getElementById('card-word').value = card.word;
        document.getElementById('card-meaning').value = card.meaning;
        document.getElementById('card-example').value = card.example;
        document.getElementById('card-modal').classList.remove('hidden');
    } catch (error) {
       // Ошибка уже показана в apiRequest
    }
}

async function deleteCard(id) {
    if (confirm('Вы уверены, что хотите удалить эту карточку?')) {
        try {
            await apiRequest(`/cards/${id}`, 'DELETE');
            loadCards();
        } catch (error) {
            // Ошибка уже показана в apiRequest
        }
    }
}


// --- Логика для страницы повторения (review.html) ---
if (window.location.pathname.endsWith('review.html')) {
    let dueCards = [];
    let currentCardIndex = 0;

    const flashcardContainer = document.getElementById('flashcard-container');
    const noCardsMessage = document.getElementById('no-cards-message');
    const flashcard = document.querySelector('.flashcard');
    
    flashcard.addEventListener('click', () => {
        flashcard.classList.toggle('flipped');
    });
    
    document.querySelectorAll('.quality-btn').forEach(button => {
        button.addEventListener('click', async () => {
            const quality = parseInt(button.dataset.quality);
            const cardId = dueCards[currentCardIndex].id;
            try {
                await apiRequest(`/cards/review/${cardId}`, 'POST', { quality });
                currentCardIndex++;
                displayCurrentCard();
            } catch (error) {
                 // Ошибка уже показана в apiRequest
            }
        });
    });
}

async function loadDueCards() {
    try {
        const reviewContainer = document.getElementById('review-container');
        dueCards = await apiRequest('/cards/due', 'GET');
        currentCardIndex = 0;
        
        if (dueCards && dueCards.length > 0) {
            document.getElementById('flashcard-container').classList.remove('hidden');
            document.getElementById('no-cards-message').classList.add('hidden');
            displayCurrentCard();
        } else {
            document.getElementById('flashcard-container').classList.add('hidden');
            document.getElementById('no-cards-message').classList.remove('hidden');
        }
    } catch (error) {
        // Ошибка уже показана в apiRequest
    }
}

function displayCurrentCard() {
    const flashcard = document.querySelector('.flashcard');
    
    if (currentCardIndex < dueCards.length) {
        const card = dueCards[currentCardIndex];
        document.getElementById('card-word-review').innerText = card.word;
        document.getElementById('card-meaning-review').innerText = card.meaning;
        document.getElementById('card-example-review').innerText = card.example || '';
        flashcard.classList.remove('flipped');
    } else {
        document.getElementById('flashcard-container').classList.add('hidden');
        document.getElementById('no-cards-message').classList.remove('hidden');
    }
}