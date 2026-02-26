export function weatherRotationPosition(): number {
    const w = document.querySelector(".weather");
    if (!w) return 0;

    const weatherPositionData = w.getAttribute("data-weather-position");
    if (!weatherPositionData) return 0;

    const weatherPosition = parseInt(weatherPositionData, 10);
    if (Number.isNaN(weatherPosition)) return 0;

    return weatherPosition;
}
