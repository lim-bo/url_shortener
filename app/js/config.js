const AppConfig = {
    API_BASE_URL: 'https://url-short-af.space/api/v1',
    API_PROVIDED_LINK_PROTOCOL: "https://",
    API_ENDPOINTS: {
        SHORTEN: '/shorten',
        STATS: '/stats'
    },
    
    UI_SETTINGS: {
        DEFAULT_ERROR_MESSAGE: 'Произошла ошибка. Пожалуйста, попробуйте позже.',
        URL_VALIDATION_MESSAGE: 'Введите корректный URL (начинается с http:// или https://)',
        EMPTY_CODE_MESSAGE: 'Введите короткий код'
    },
    
    DEV_SETTINGS: {
        DEBUG_MODE: true,
        MOCK_API: false
    }
};

window.AppConfig = AppConfig;