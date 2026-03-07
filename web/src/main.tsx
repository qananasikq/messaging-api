import React from "react";
import { createRoot } from "react-dom/client";
import App from "./App.tsx";

import "./styles.css";

// Получаем корневой элемент
const rootElement = document.getElementById("root");

if (!rootElement) {
  throw new Error(
    "Root element not found. Make sure you have <div id=\"root\"></div> in index.html"
  );
}

// Создаём root и рендерим приложение
const root = createRoot(rootElement);

root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);

// Опционально: для hot module replacement в dev-режиме (Vite / Create React App)
if (import.meta.env.DEV && import.meta.hot) {
  import.meta.hot.accept("./App.tsx", () => {
    // HMR уже обрабатывается автоматически в большинстве случаев
  });
}