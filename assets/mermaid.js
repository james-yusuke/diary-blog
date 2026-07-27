const sourceBlocks = document.querySelectorAll("pre > code.language-mermaid");

if (sourceBlocks.length > 0) {
  const { default: mermaid } = await import("https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs");
  mermaid.initialize({
    startOnLoad: false,
    securityLevel: "strict",
    theme: window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "default",
    flowchart: { htmlLabels: true, useMaxWidth: true },
  });

  let diagramNumber = 0;
  for (const sourceBlock of sourceBlocks) {
    const source = sourceBlock.textContent;
    const diagram = document.createElement("div");
    diagram.className = "mermaid";
    diagram.setAttribute("role", "img");
    diagram.setAttribute("aria-label", "Mermaid diagram");
    sourceBlock.parentElement.replaceWith(diagram);

    try {
      const { svg, bindFunctions } = await mermaid.render(`mermaid-${diagramNumber++}`, source);
      diagram.innerHTML = svg;
      bindFunctions?.(diagram);
    } catch (error) {
      console.warn("Could not render Mermaid diagram", error);
      const fallback = document.createElement("pre");
      const fallbackCode = document.createElement("code");
      fallbackCode.className = "language-mermaid";
      fallbackCode.textContent = source;
      fallback.append(fallbackCode);
      diagram.replaceWith(fallback);
    }
  }
}
