import { useState } from "react";
import { login, register } from "../api";

type AuthMode = "login" | "register";

type LoginProps = {
  onAuth: (token: string, username: string) => void;
};

export default function Login({ onAuth }: LoginProps) {
  const [mode, setMode] = useState<AuthMode>("login");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const isLogin = mode === "login";

  async function handleSubmit() {
    setError("");

    const trimmedUsername = username.trim();
    const trimmedPassword = password.trim();

    if (!trimmedUsername || !trimmedPassword) {
      setError("Введите имя пользователя и пароль");
      return;
    }

    setIsSubmitting(true);

    try {
      const apiCall = isLogin ? login : register;
      const response = await apiCall(trimmedUsername, trimmedPassword);

      onAuth(response.token, response.username);

      // можно очистить форму после успешного входа/регистрации
      setUsername("");
      setPassword("");
    } catch (err: any) {
      const status = err?.response?.status;

      if (status === 409) {
        setError("Пользователь с таким именем уже существует. Попробуйте войти.");
      } else if (status === 401) {
        setError("Неверное имя пользователя или пароль");
      } else if (status === 400) {
        setError(
          "Имя пользователя — 3–32 символа, пароль — 8–128 символов"
        );
      } else {
        setError(
          isLogin ? "Не удалось войти" : "Не удалось создать аккаунт"
        );
      }
    } finally {
      setIsSubmitting(false);
    }
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === "Enter") {
      e.preventDefault();
      void handleSubmit();
    }
  }

  return (
    <div className="authPage">
      <div className="authHero">
        <div className="authBadge">Go · PostgreSQL · Redis · WebSocket</div>
        <h1>Messaging App</h1>
        <p>
          Чат с поддержкой личных и групповых диалогов, уведомлениями о непрочитанных
          сообщениях, WebSocket в реальном времени и красивым интерфейсом.
        </p>
      </div>

      <div className="authCard">
        <div className="authTabs">
          <button
            type="button"
            className={isLogin ? "active" : ""}
            onClick={() => setMode("login")}
            disabled={isSubmitting}
          >
            Вход
          </button>
          <button
            type="button"
            className={!isLogin ? "active" : ""}
            onClick={() => setMode("register")}
            disabled={isSubmitting}
          >
            Регистрация
          </button>
        </div>

        <div className="authFields">
          <label>
            Имя пользователя
            <input
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="например, alex или dima"
              autoFocus
              autoComplete="username"
            />
          </label>

          <label>
            Пароль
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="минимум 8 символов"
              autoComplete={isLogin ? "current-password" : "new-password"}
            />
          </label>
        </div>

        {error && <div className="formError">{error}</div>}

        <button
          type="button"
          className="primaryButton fullWidth"
          onClick={handleSubmit}
          disabled={isSubmitting}
        >
          {isSubmitting
            ? "Подождите..."
            : isLogin
              ? "Войти"
              : "Зарегистрироваться"}
        </button>

        <div className="authFooter">
          {isLogin ? (
            <p>
              Нет аккаунта?{" "}
              <button
                type="button"
                className="linkButton"
                onClick={() => setMode("register")}
              >
                Зарегистрироваться
              </button>
            </p>
          ) : (
            <p>
              Уже есть аккаунт?{" "}
              <button
                type="button"
                className="linkButton"
                onClick={() => setMode("login")}
              >
                Войти
              </button>
            </p>
          )}
        </div>
      </div>
    </div>
  );
}