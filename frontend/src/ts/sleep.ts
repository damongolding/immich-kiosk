function sleep(turnOn: boolean): void {
  const fk = window.fully;

  if (turnOn) {
    document.body.classList.add("sleep");
    if (typeof fk !== "undefined" && fk.getScreenOn()) {
      fk.showToast("Entering sleep mode");
      setTimeout(() => fk.turnScreenOff(), 4 * 1000);
    }
    return;
  }

  document.body.classList.remove("sleep");
  if (typeof fk !== "undefined" && !fk.getScreenOn()) {
    fk.turnScreenOn();
    fk.showToast("Exited sleep mode");
  }
}

export { sleep };
