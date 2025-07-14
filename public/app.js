const apiBaseUrl = 'http://localhost:8080';
let token = localStorage.getItem('token') || null;

function showAuthSection() {
    const authSection = document.getElementById('auth-section');
    const flashcardsSection = document.getElementById('flashcards-section');
    if (authSection && flashcardsSection) {
        authSection.style.display = 'block';
        flashcardsSection.style.display = 'none';
    } else {
        console.error('Auth or Flashcards section not found');
    }
}

function showFlashcardsSection() {
    const authSection = document.getElementById('auth-section');
    const flashcardsSection = document.getElementById('flashcards-section');
    if (authSection && flashcardsSection) {
        authSection.style.display = 'none';
        flashcardsSection.style.display = 'block';
        loadFlashcards();
        loadDueFlashcards();
    } else {
        console.error('Auth or Flashcards section not found');
    }
}

function logout() {
    localStorage.removeItem('token');
    token = null;
    showAuthSection();
}

if (token) {
    showFlashcardsSection();
} else {
    showAuthSection();
}

// Обработчик формы регистрации
document.getElementById('register-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const email = document.getElementById('register-email').value;
    const password = document.getElementById('register-password').value;

    try {
        const response = await fetch(`${apiBaseUrl}/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });
        const data = await response.json();
        if (response.ok) {
            alert('Registration successful! Please log in.');
        } else {
            alert(`Error: ${data.error}`);
        }
    } catch (error) {
        alert(`Error: ${error.message}`);
    }
});

// Обработчик формы логина
document.getElementById('login-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;

    try {
        const response = await fetch(`${apiBaseUrl}/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });
        const data = await response.json();
        if (response.ok) {
            token = data.token;
            localStorage.setItem('token', token);
            showFlashcardsSection();
        } else {
            alert(`Error: ${data.error}`);
        }
    } catch (error) {
        alert(`Error: ${error.message}`);
    }
});

// Обработчик формы добавления карточки
document.getElementById('add-flashcard-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const word = document.getElementById('word').value;
    const meaning = document.getElementById('meaning').value;
    const example = document.getElementById('example').value;

    try {
        const response = await fetch(`${apiBaseUrl}/cards`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({ word, meaning, example })
        });
        const data = await response.json();
        if (response.ok) {
            alert('Flashcard added!');
            document.getElementById('add-flashcard-form').reset();
            loadFlashcards();
        } else {
            alert(`Error: ${data.error}`);
        }
    } catch (error) {
        alert(`Error: ${error.message}`);
    }
});

// Загрузка всех карточек
async function loadFlashcards() {
    const tbody = document.querySelector('#flashcards-table tbody');
    if (!tbody) {
        console.error('Flashcards table tbody not found');
        alert('Error: Flashcards table not found. Please refresh the page.');
        return;
    }

    try {
        const response = await fetch(`${apiBaseUrl}/cards`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        const cards = await response.json();
        console.log('Flashcards response:', cards); // Отладка
        tbody.innerHTML = '';
        if (Array.isArray(cards) && cards.length > 0) {
            cards.forEach(card => {
                const row = document.createElement('tr');
                row.innerHTML = `
                    <td>${card.word}</td>
                    <td>${card.meaning}</td>
                    <td>${card.example || ''}</td>
                    <td>${new Date(card.next_review).toLocaleString()}</td>
                    <td>
                        <button onclick="deleteFlashcard(${card.id})">Delete</button>
                    </td>
                `;
                tbody.appendChild(row);
            });
        } else {
            tbody.innerHTML = '<tr><td colspan="5">No flashcards available</td></tr>';
        }
    } catch (error) {
        console.error('Error loading flashcards:', error);
        alert(`Error: ${error.message}`);
    }
}

// Удаление карточки
async function deleteFlashcard(id) {
    try {
        const response = await fetch(`${apiBaseUrl}/cards/${id}`, {
            method: 'DELETE',
            headers: { 'Authorization': `Bearer ${token}` }
        });
        const data = await response.json();
        if (response.ok) {
            alert('Flashcard deleted!');
            loadFlashcards();
        } else {
            alert(`Error: ${data.error}`);
        }
    } catch (error) {
        alert(`Error: ${error.message}`);
    }
}

// Загрузка карточек для повторения
async function loadDueFlashcards() {
    const container = document.getElementById('due-flashcards');
    const noDueMessage = document.getElementById('no-due-message');

    if (!container || !noDueMessage) {
        console.error('Due flashcards container or no-due-message not found');
        alert('Error: Due flashcards section not found. Please refresh the page.');
        return;
    }

    try {
        const response = await fetch(`${apiBaseUrl}/cards/due`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        const cards = await response.json();
        console.log('Due flashcards response:', cards); // Отладка
        container.innerHTML = '';
        noDueMessage.style.display = 'none';
        if (Array.isArray(cards) && cards.length > 0) {
            cards.forEach(card => {
                const div = document.createElement('div');
                div.className = 'flashcard';
                div.innerHTML = `
                    <p><strong>Word:</strong> ${card.word}</p>
                    <p style="display: none;" class="meaning"><strong>Meaning:</strong> ${card.meaning}</p>
                    <button onclick="this.previousElementSibling.style.display='block';">Show Meaning</button>
                    <p><strong>Example:</strong> ${card.example || ''}</p>
                    <p><strong>Next Review:</strong> ${new Date(card.next_review).toLocaleString()}</p>
                    <p><strong>Repetitions:</strong> ${card.repetitions}</p>
                    <p><strong>Easiness Factor:</strong> ${card.ef.toFixed(2)}</p>
                    <p>Rate your recall (0-5):</p>
                    <input type="number" min="0" max="5" id="quality-${card.id}">
                    <button onclick="reviewFlashcard(${card.id})">Submit Review</button>
                    <hr>
                `;
                container.appendChild(div);
            });
        } else {
            noDueMessage.style.display = 'block';
        }
    } catch (error) {
        console.error('Error loading due flashcards:', error);
        alert(`Error: ${error.message}`);
        container.innerHTML = '';
        noDueMessage.style.display = 'block';
    }
}

// Отправка оценки повторения
async function reviewFlashcard(id) {
    const quality = parseInt(document.getElementById(`quality-${id}`).value);
    if (isNaN(quality) || quality < 0 || quality > 5) {
        alert('Quality must be between 0 and 5');
        return;
    }

    try {
        const response = await fetch(`${apiBaseUrl}/cards/review/${id}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({ quality })
        });
        const data = await response.json();
        if (response.ok) {
            alert('Review submitted!');
            loadDueFlashcards();
        } else {
            alert(`Error: ${data.error}`);
        }
    } catch (error) {
        console.error('Error submitting review:', error);
        alert(`Error: ${error.message}`);
    }
}

// Загрузка карточек при открытии страницы
document.addEventListener('DOMContentLoaded', () => {
    if (token) {
        showFlashcardsSection();
    } else {
        showAuthSection();
    }
});