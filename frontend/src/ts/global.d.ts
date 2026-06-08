import type { VideoMuteStatus } from "./mute";

type ImmichKioskBrowserApi = {
    video: {
        getStatus: () => VideoMuteStatus;
        getMuted: () => boolean;
        setMuted: (muted: boolean) => boolean;
        toggleMuted: () => boolean;
    };
};

declare global {
    interface Window {
        immichKiosk?: ImmichKioskBrowserApi;
    }
}
