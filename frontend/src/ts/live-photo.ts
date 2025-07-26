enum LivePhoto {
    ATTRIBUTE = "data-live-photo",
    TOP = "top",
    BOTTOM = "bottom",
    FADE_CLASS = "live-photo-fade",
}

export function livePhoto(delay: number) {
    document.addEventListener(
        "ended",
        (event) => {
            const target = event.target;
            if (!(target instanceof HTMLVideoElement)) return;
            if (!target.hasAttribute(LivePhoto.ATTRIBUTE)) return;

            const role = target.dataset.livePhotoRole;
            const group = target.dataset.group;
            if (!role || !group) return;

            const isTop = role === LivePhoto.TOP;

            const otherRole = isTop ? LivePhoto.BOTTOM : LivePhoto.TOP;

            const selector = `video[${LivePhoto.ATTRIBUTE}][data-group="${group}"][data-live-photo-role="${otherRole}"]`;
            const otherVideo = document.querySelector(selector);

            if (!(otherVideo instanceof HTMLVideoElement)) return;

            otherVideo.currentTime = 0;

            if (isTop) {
                target.classList.add(LivePhoto.FADE_CLASS);
            } else {
                otherVideo.classList.remove(LivePhoto.FADE_CLASS);
            }

            setTimeout(() => {
                if (!otherVideo.isConnected) return;
                otherVideo.play();
            }, delay);
        },
        true,
    );
}
