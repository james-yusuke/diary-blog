(() => {
  const dialog = document.querySelector("#command-palette");
  const trigger = document.querySelector("[data-command-trigger]");
  const input = document.querySelector("#command-query");
  const items = Array.from(document.querySelectorAll("[data-command-item]"));
  const close = document.querySelector("[data-command-close]");
  if (!dialog || !trigger || !input || !close) return;

  let activeIndex = 0;
  const visibleItems = () => items.filter((item) => !item.hidden);
  const setActive = (index) => {
    const visible = visibleItems();
    if (visible.length === 0) return;
    activeIndex = (index + visible.length) % visible.length;
    visible.forEach((item, itemIndex) => item.setAttribute("aria-selected", String(itemIndex === activeIndex)));
  };
  const filter = () => {
    const query = input.value.trim().toLocaleLowerCase();
    items.forEach((item) => {
      item.hidden = !item.textContent.toLocaleLowerCase().includes(query);
    });
    activeIndex = 0;
    setActive(0);
  };
  const open = () => {
    if (!dialog.open) dialog.showModal();
    input.value = "";
    filter();
    input.focus({ preventScroll: true });
  };
  const closePalette = () => {
    if (!dialog.open) return;
    dialog.close();
    trigger.focus({ preventScroll: true });
  };

  trigger.addEventListener("click", open);
  close.addEventListener("click", closePalette);
  input.addEventListener("input", filter);
  input.addEventListener("keydown", (event) => {
    if (event.key === "ArrowDown") {
      event.preventDefault();
      setActive(activeIndex + 1);
    }
    if (event.key === "ArrowUp") {
      event.preventDefault();
      setActive(activeIndex - 1);
    }
    if (event.key === "Enter" && input.value.trim() === "") {
      const item = visibleItems()[activeIndex];
      if (item) item.click();
    }
  });
  document.addEventListener("keydown", (event) => {
    const target = event.target;
    const isTyping = target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement || target instanceof HTMLSelectElement;
    if ((event.metaKey || event.ctrlKey) && event.key.toLocaleLowerCase() === "k") {
      event.preventDefault();
      open();
    }
    if (event.key === "/" && !isTyping && !dialog.open) {
      event.preventDefault();
      open();
    }
    if (event.key === "Escape" && dialog.open) {
      event.preventDefault();
      closePalette();
    }
  });
})();
