import { useEffect, useMemo, useState, useCallback } from "react";
import Login from "./components/Login";
import DialogList from "./components/DialogList";
import Chat from "./components/Chat";
import { setToken } from "./api";

type Session = {
  token: string;
  userId: string;
  username: string;
};

function parseJwtPayload(token: string): { uid?: string; usr?: string } | null {
  try {
    const [, payloadBase64] = token.split(".");
    const json = atob(payloadBase64.replace(/-/g, "+").replace(/_/g, "/"));
    return JSON.parse(json);
  } catch {
    return null;
  }
}

function restoreSession(token: string, fallbackUsername?: string): Session | null {
  const payload = parseJwtPayload(token);
  if (!payload?.uid) return null;

  return {
    token,
    userId: payload.uid,
    username: fallbackUsername || payload.usr || "Anonymous",
  };
}

export default function App() {
  const [session, setSession] = useState<Session | null>(null);
  const [selectedDialogId, setSelectedDialogId] = useState<string | null>(null);
  const [dialogsVersion, setDialogsVersion] = useState(0);

  // Восстановление сессии при монтировании
  useEffect(() => {
    const savedToken = localStorage.getItem("token");
    if (!savedToken) return;

    const savedUsername = localStorage.getItem("username") ?? undefined;
    const restored = restoreSession(savedToken, savedUsername);

    if (restored) {
      setToken(restored.token);
      setSession(restored);
    } else {
      // Токен битый → чистим хранилище
      localStorage.removeItem("token");
      localStorage.removeItem("username");
    }
  }, []);

  const handleAuth = (token: string, username: string) => {
    const session = restoreSession(token, username);
    if (!session) return;

    localStorage.setItem("token", token);
    localStorage.setItem("username", session.username);
    setToken(token);
    setSession(session);
  };

  const handleLogout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("username");
    setToken(null);
    setSession(null);
    setSelectedDialogId(null);
  };

  // увеличиваем версию, чтобы список диалогов обновился
  // при этом Chat не размонтируется (версия не участвует в ключе).
  const handleDialogsChanged = useCallback(() => {
    setDialogsVersion((prev) => prev + 1);
  }, []);

  // Chat должен перемонтироваться только при выборе другого диалога.
  const chatKey = useMemo(
    () => selectedDialogId ?? "no-dialog",
    [selectedDialogId]
  );

  if (!session) {
    return <Login onAuth={handleAuth} />;
  }

  return (
    <div className="layoutShell">
      <div className="appFrame">
        <aside className="sidebar">
          <DialogList
            currentUserId={session.userId}
            username={session.username}
            selectedDialogId={selectedDialogId}
            dialogsVersion={dialogsVersion}
            onOpen={setSelectedDialogId}
            onLogout={handleLogout}
            onDialogsChanged={handleDialogsChanged}
          />
        </aside>

        <main className="chatPanel">
          {selectedDialogId ? (
            <Chat
              key={chatKey}
              token={session.token}
              currentUserId={session.userId}
              dialogId={selectedDialogId}
              onDialogDeleted={() => {
                setSelectedDialogId(null);
                handleDialogsChanged();
              }}
              onDialogUpdated={handleDialogsChanged}
            />
          ) : (
            <div className="emptyState">
              <div className="emptyStateBadge">Messaging App</div>
              <h2>Добро пожаловать!</h2>
              <p>
                Выберите чат из списка слева или создайте новый.
                <br />
                Можно также открыть любой диалог по его ID.
              </p>
            </div>
          )}
        </main>
      </div>
    </div>
  );
}