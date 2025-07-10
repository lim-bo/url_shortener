// Элементы DOM
const shortenBtn = document.getElementById('shortenBtn');
const statsBtn = document.getElementById('statsBtn');
const shortenResult = document.getElementById('shortenResult');
const statsResult = document.getElementById('statsResult');

if (!window.AppConfig) {
    console.error('Конфигурация приложения не загружена!');
}
// Обработчик для кнопки сокращения ссылки
shortenBtn.addEventListener('click', async () => {
    const urlInput = document.getElementById('originalUrl').value.trim();
    
    if (!isValidUrl(urlInput)) {
        showResult(shortenResult, AppConfig.UI_SETTINGS.URL_VALIDATION_MESSAGE, 'error');
        return;
    }

    try {
        // Отправка запроса к API
        const response = await fetch(`${AppConfig.API_BASE_URL}${AppConfig.API_ENDPOINTS.SHORTEN}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ link: urlInput })
        });

        // Обработка ответа
        if (response.ok) {
            const data = await response.json();
            const shortUrl = data.link;
            showResult(shortenResult, `
                Ваша сокращенная ссылка: 
                <a href="${AppConfig.API_PROVIDED_LINK_PROTOCOL}${shortUrl}" class="link" target="_blank">${AppConfig.API_PROVIDED_LINK_PROTOCOL}${shortUrl}</a>
            `, 'success');
        } else {
            const error = await response.text();
            showResult(shortenResult, `Ошибка: ${error}`, 'error');
        }
   } catch (err) {
        showResult(shortenResult, 
            AppConfig.DEV_SETTINGS.DEBUG_MODE ? 
            `Сетевая ошибка: ${err.message}` : 
            AppConfig.UI_SETTINGS.DEFAULT_ERROR_MESSAGE, 
            'error'
        );
    }
});

// Обработчик для кнопки получения статистики
statsBtn.addEventListener('click', async () => {
    const shortCode = document.getElementById('shortCode').value.trim();
    
    if (!shortCode) {
        showResult(statsResult, AppConfig.UI_SETTINGS.EMPTY_CODE_MESSAGE, 'error');
        return;
    }

    try {
         const response = await fetch(
            `${AppConfig.API_BASE_URL}${AppConfig.API_ENDPOINTS.STATS}/${encodeURIComponent(shortCode)}`
        );
        if (response.ok) {
            const stats = await response.json();
            showResult(statsResult, `
                <div class="stats-card">
                    <div class="stat-item">
                        <span class="label">Оригинальная ссылка:</span><br>
                        <a href="${stats.link}" class="value link" target="_blank">${stats.link}</a>
                    </div>
                    <div class="stat-item">
                        <span class="label">Короткий код:</span>
                        <span class="value">${stats.code}</span>
                    </div>
                    <div class="stat-item">
                        <span class="label">Количество переходов:</span>
                        <span class="value">${stats.clicks}</span>
                    </div>
                </div>
            `, 'success');
        } else {
            const error = await response.text();
            showResult(statsResult, `Ошибка: ${error}`, 'error');
        }
    } catch (err) {
       showResult(shortenResult, 
            AppConfig.DEV_SETTINGS.DEBUG_MODE ? 
            `Сетевая ошибка: ${err.message}` : 
            AppConfig.UI_SETTINGS.DEFAULT_ERROR_MESSAGE, 
            'error'
        );
    }
});

// Функция валидации URL
function isValidUrl(url) {
    try {
        new URL(url);
        return true;
    } catch {
        return false;
    }
}

// Функция отображения результатов
function showResult(element, message, type) {
    element.innerHTML = message;
    element.className = `result ${type}`;
    element.style.display = 'block';
}