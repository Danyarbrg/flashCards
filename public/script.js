const API_URL = 'http://localhost:8080';; 

let dueCards = [];
let currentCardIndex = 0;
let currentSortBy = 'created';
let currentSortOrder = 'asc';
let allUserTags = [];

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
        if (response.status === 204 || response.headers.get("Content-Length") === "0") {
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

function toggleForms() {
    document.getElementById('login-form').classList.toggle('hidden');
    document.getElementById('register-form').classList.toggle('hidden');
}

async function login(event) {
    event.preventDefault();
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;
    try {
        const data = await apiRequest('/login', 'POST', { email, password });
        localStorage.setItem('token', data.token);
        window.location.href = '/cards.html';
    } catch (error) {}
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
    } catch (error) {}
}

function initializeCardsPage() {
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
        if (event.target == modal) modal.classList.add('hidden');
    }

        cardForm.onsubmit = async (event) => {
        event.preventDefault();
        const id = cardIdInput.value;
        const cardData = {
            word: document.getElementById('card-word').value,
            meaning: document.getElementById('card-meaning').value,
            example: document.getElementById('card-example').value,
            tags: document.getElementById('card-tags').value,
        };

        try {
            if (id) {
                await apiRequest(`/cards/${id}`, 'PUT', cardData);
            } else {
                await apiRequest('/cards', 'POST', cardData);
            }
            modal.classList.add('hidden');
            loadCards();
            loadUserTags();
        } catch (error) {}
    };

    document.getElementById('sort-by').addEventListener('change', (e) => {
        currentSortBy = e.target.value;
        loadCards();
    });
    document.getElementById('sort-order').addEventListener('change', (e) => {
        currentSortOrder = e.target.value;
        loadCards();
    });

    const tagFilterInput = document.getElementById('tag-filter');
    const clearFilterBtn = document.getElementById('clear-filter-btn');

    tagFilterInput.addEventListener('keypress', function(event) {
        if (event.key === 'Enter') {
            event.preventDefault();
            loadCards();
        }
    });

    clearFilterBtn.addEventListener('click', () => {
        tagFilterInput.value = '';
        loadCards();
    });

    loadCards();
    loadUserTags();
}


async function loadCards() {
    try {
        const tagFilter = document.getElementById('tag-filter').value;
        // Формируем URL с учетом фильтра
        let endpoint = `/cards?sort=${currentSortBy}&order=${currentSortOrder}`;
        if (tagFilter) {
            endpoint += `&tag=${encodeURIComponent(tagFilter)}`;
        }

        const cards = await apiRequest(endpoint, 'GET');
        const container = document.getElementById('cards-container');
        container.innerHTML = '';
        if (cards && cards.length > 0) {
            cards.forEach(card => {
                const cardElement = document.createElement('div');
                cardElement.className = 'card';

                // Создаем контейнер для тегов
                let tagsHTML = '';
                if (card.tags) {
                    const tagsArray = card.tags.split(',').map(tag => tag.trim());
                    tagsHTML = `<div class="tags-container">
                        ${tagsArray.map(tag => `<span class="tag-item">${tag}</span>`).join('')}
                    </div>`;
                }
                
                cardElement.innerHTML = `
                    <div>
                        <h3>${card.word}</h3>
                        <p>${card.meaning}</p>
                        ${card.example ? `<em>${card.example}</em>` : ''}
                        ${tagsHTML} 
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
    } catch (error) {}
}

async function editCard(id) {
    try {
        const card = await apiRequest(`/cards/${id}`, 'GET');
        document.getElementById('modal-title').innerText = "Редактировать карточку";
        document.getElementById('card-id').value = card.id;
        document.getElementById('card-word').value = card.word;
        document.getElementById('card-meaning').value = card.meaning;
        document.getElementById('card-example').value = card.example;
        document.getElementById('card-tags').value = card.tags;
        document.getElementById('card-modal').classList.remove('hidden');
    } catch (error) {}
}

async function loadUserTags() {
    try {
        allUserTags = await apiRequest('/cards/tags', 'GET') || [];
        const datalist = document.getElementById('tag-suggestions');
        datalist.innerHTML = '';
        allUserTags.forEach(tag => {
            const option = document.createElement('option');
            option.value = tag;
            datalist.appendChild(option);
        });
    } catch (error) {
        console.error("Failed to load user tags:", error);
    }
}

async function deleteCard(id) {
    if (confirm('Вы уверены, что хотите удалить эту карточку?')) {
        try {
            await apiRequest(`/cards/${id}`, 'DELETE');
            loadCards();
        } catch (error) {}
    }
}

function initializeReviewPage() {
    const flashcard = document.querySelector('.flashcard');
    
    if (flashcard) {
        flashcard.addEventListener('click', () => {
            flashcard.classList.toggle('flipped');
        });
    }
    
    document.querySelectorAll('.quality-btn').forEach(button => {
        button.addEventListener('click', async () => {
            const quality = parseInt(button.dataset.quality);
            const cardId = dueCards[currentCardIndex].id;
            try {
                await apiRequest(`/cards/review/${cardId}`, 'POST', { quality });
                currentCardIndex++;
                displayCurrentCard();
            } catch (error) {}
        });
    });
    
    loadDueCards();
}

async function loadDueCards() {
    try {
        dueCards = await apiRequest('/cards/due', 'GET');
        currentCardIndex = 0;
        displayCurrentCard();
    } catch (error) {}
}

function displayCurrentCard() {
    const flashcardContainer = document.getElementById('flashcard-container');
    const noCardsMessage = document.getElementById('no-cards-message');
    const flashcard = document.querySelector('.flashcard');

    if (dueCards && currentCardIndex < dueCards.length) {
        flashcardContainer.classList.remove('hidden');
        noCardsMessage.classList.add('hidden');
        
        const card = dueCards[currentCardIndex];
        document.getElementById('card-word-review').innerText = card.word;
        document.getElementById('card-meaning-review').innerText = card.meaning;
        document.getElementById('card-example-review').innerText = card.example || '';
        
        if (flashcard.classList.contains('flipped')) {
            flashcard.classList.remove('flipped');
        }
    } else {
        flashcardContainer.classList.add('hidden');
        noCardsMessage.classList.remove('hidden');
    }
}