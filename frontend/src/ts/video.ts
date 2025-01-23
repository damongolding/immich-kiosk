import htmx from "htmx.org";
import { nextAsset } from "./polling";

let video: HTMLVideoElement | null = null;
let progress: HTMLElement | null = null;
let animationFrameId: number | null = null;

function videoHandler(id: string): void {
  if (!id) return;

  video = document.getElementById(id) as HTMLVideoElement;

  if (!video) {
    console.error("Video element not found");
    return;
  }

  const progressBarElement = htmx.find(".progress--bar") as HTMLElement | null;
  progress = progressBarElement;
  progressBarElement?.classList.remove("progress--bar-paused");

  // Start the smooth progress update
  startSmoothProgress();

  video.addEventListener(
    "ended",
    () => {
      if (animationFrameId !== null) {
        cancelAnimationFrame(animationFrameId);
      }
      videoCleanup();
      nextAsset();
    },
    false,
  );
}

function startSmoothProgress(): void {
  function updateProgressSmooth() {
    if (!video || !progress) return;

    const percentage = (video.currentTime / video.duration) * 100;
    progress.style.width = `${percentage}%`;

    animationFrameId = requestAnimationFrame(updateProgressSmooth);
  }

  updateProgressSmooth();
}

// Clean up function when needed
function videoCleanup(): void {
  if (animationFrameId !== null) {
    cancelAnimationFrame(animationFrameId);
  }
}

export { videoHandler };
