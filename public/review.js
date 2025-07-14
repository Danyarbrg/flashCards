const apiBaseUrl = 'http://localhost:8080';
const token = localStorage.getItem('token');

if (!token) {
    window.location.href = '/'; // Перенаправление на главную, если не авторизован
}

function logout() {
    localStorage.removeItem('token');
    window.location.href = '/';
}

// Загрузка карточек для повторения
async function loadDueFlashcards() {
    const sortBy = document.getElementById('sort-by').value;
    const sortOrder = document.getElementById('sort-order').value;
    const container = document.getElementById('due-flashcards');
    const noDueMessage = document.getElementById('no-due-message');

    if (!container || !noDueMessage) {
        console.error('Required DOM elements not found');
        alert('Error: Page elements not found. Please refresh the page.');
        return;
    }

    try {
        const response = await fetch(`${apiBaseUrl}/cards/due`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        let cards = await response.json();
        console.log('Due flashcards response:', cards); // Отладка

        if (!response.ok) {
            alert(`Error: ${cards.error}`);
            container.innerHTML = '';
            noDueMessage.style.display = 'block';
            return;
        }

        // Сортировка на клиенте
        if (Array.isArray(cards)) {
            cards.sort((a, b) => {
                let valA = a[sortBy];
                let valB = b[sortBy];

                if (sortBy === 'next_review') {
                    valA = new Date(valA).getTime();
                    valB = new Date(valB).getTime();
                } else if (sortBy === 'word') {
                    valA = valA.toLowerCase();
                    valB = valB.toLowerCase();
                }

                if (sortOrder === 'asc') {
                    return valA > valB ? 1 : -1;
                } else {
                    return valA < valB ? 1 : -1;
                }
            });
        }

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

// Обработчик формы сортировки
document.getElementById('sort-form').addEventListener('submit', (e) => {
    e.preventDefault();
    loadDueFlashcards();
});

// Загрузка карточек при открытии страницы
document.addEventListener('DOMContentLoaded', () => {
    loadDueFlashcards();
});