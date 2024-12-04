import { cleanupFrames } from "./kiosk";

const IMAGE_LOAD_TIMEOUT_MS = 10000 as const;

/**
 * Represents the result of loading an image
 */
interface ImageLoadResult {
  success: boolean;
  img: HTMLImageElement;
}

/**
 * Handles a new frame element by loading all images within it and marking it as loaded
 * @param target - The target element to handle (either a frame or container of frames)
 * @returns Promise that resolves when frame is handled
 * @throws No errors thrown, but logs warnings for missing frames or failed image loads
 */
async function handleNewFrame(target: HTMLElement | null): Promise<void> {
  if (!target) {
    console.warn("Target frame is null");
    return;
  }
  const frames: HTMLElement[] = target.classList.contains("frame")
    ? [target]
    : Array.from(target.querySelectorAll<HTMLElement>(".frame"));

  const lastFrame: HTMLElement = Array.from(frames).pop() as HTMLElement;
  if (!lastFrame) {
    console.warn("No frame elements found");
    return;
  }
  const images = Array.from(
    lastFrame.querySelectorAll<HTMLImageElement>("img"),
  );

  if (images.length === 0) {
    lastFrame.classList.add("loaded");
    await cleanupFrames();
    return;
  }

  const checkImageRendered = (img: HTMLImageElement): Promise<boolean> => {
    return new Promise((resolve) => {
      let attempts = 0;
      const maxAttempts = 3;

      const check = () => {
        const styles = window.getComputedStyle(img);
        const bounds = img.getBoundingClientRect();

        // More comprehensive checks
        const hasSize =
          (bounds.width > 0 && bounds.height > 0) || // Check actual rendered size
          (img.clientWidth > 0 && img.clientHeight > 0) ||
          (img.naturalWidth > 0 && img.naturalHeight > 0);

        const isRendered =
          img.complete &&
          hasSize &&
          styles.display !== "none" &&
          styles.visibility !== "hidden" &&
          styles.opacity !== "0";

        if (isRendered || attempts >= maxAttempts) {
          resolve(isRendered);
        } else {
          attempts++;
          requestAnimationFrame(check);
        }
      };

      requestAnimationFrame(check);
    });
  };

  const imagePromises = images.map(async (img): Promise<ImageLoadResult> => {
    try {
      if (img.complete) {
        const isRendered = await checkImageRendered(img);
        return { success: isRendered, img };
      }

      img.loading = "eager";

      const result = await Promise.race([
        new Promise<ImageLoadResult>((resolve) => {
          const handleLoad = async () => {
            img.removeEventListener("load", handleLoad);
            img.removeEventListener("error", handleError);
            const isRendered = await checkImageRendered(img);
            resolve({ success: isRendered, img });
          };

          const handleError = () => {
            img.removeEventListener("load", handleLoad);
            img.removeEventListener("error", handleError);
            resolve({ success: false, img });
          };

          img.addEventListener("load", handleLoad);
          img.addEventListener("error", handleError);
        }),

        new Promise<ImageLoadResult>((resolve) =>
          setTimeout(
            () => resolve({ success: false, img }),
            IMAGE_LOAD_TIMEOUT_MS,
          ),
        ),
      ]);

      return result;
    } catch (error) {
      console.error(`Error loading image: ${img.src}`, error);
      return { success: false, img };
    }
  });

  const results = await Promise.all(imagePromises);

  const failedImages = results.filter((r) => !r.success);
  if (failedImages.length > 0) {
    console.warn(
      `Failed to load ${failedImages.length} images:`,
      failedImages.map((r) => r.img.src),
    );
  }

  await new Promise((resolve) => setTimeout(resolve, 50));
  await new Promise((resolve) => requestAnimationFrame(resolve));

  lastFrame.classList.add("loaded");

  await cleanupFrames();
}

export { handleNewFrame };
