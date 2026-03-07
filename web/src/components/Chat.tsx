
import { useEffect, useMemo, useRef, useState } from "react";
import { deleteDialog, Dialog, getDialog, getMessages, Message, sendMessage } from "../api";
import { createDialogSocket } from "../ws";

function formatTime(isoString: string): string {
  return new Date(isoString).toLocaleString("ru-RU", {
    dateStyle: "short",
    timeStyle: "short",
  });
}

function getDialogTitle(dialog: Dialog, currentUserId: string): string {
  if (dialog.name) return dialog.name;

  const otherParticipants = dialog.participants?.filter((p) => p.id !== currentUserId) ?? [];

  if (otherParticipants.length === 0) {
    return dialog.type === "group" ? "Групповой чат" : "Личный чат";
  }

  return otherParticipants.map((p) => p.username).join(", ");
}

type ChatProps = {
  token: string;
  currentUserId: string;
  dialogId: string;
  onDialogDeleted: () => void;
  onDialogUpdated: () => void;
};

export default function Chat({
  token,
  currentUserId,
  dialogId,
  onDialogDeleted,
  onDialogUpdated,
}: ChatProps) {
  const [dialog, setDialog] = useState<Dialog | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [text, setText] = useState("");
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(true);

  const messagesEndRef = useRef<HTMLDivElement>(null);
  const socketRef = useRef<WebSocket | null>(null);

  // хелпер добавляет сообщение только один раз по ID
  const addMessage = (msg: Message) => {
    setMessages((prev) => {
      if (prev.some((m) => m.id === msg.id)) return prev;
      return [...prev, msg];
    });
  };

  useEffect(() => {
    let isCancelled = false;

    async function loadDialogAndMessages() {
      setIsLoading(true);
      setError("");

      try {
        const [dialogData, messagesData] = await Promise.all([getDialog(dialogId), getMessages(dialogId)]);

        if (isCancelled) return;

        // Сообщения приходят от старых к новым → разворачиваем для отображения снизу
        setMessages([...messagesData].reverse());
        setDialog(dialogData);

        // Подключаем WebSocket после того, как данные загружены.
        const connectSocket = () => {
          if (socketRef.current) {
            socketRef.current.close();
            socketRef.current = null;
          }

          const sock = createDialogSocket(token, dialogId);
          socketRef.current = sock;

          sock.onmessage = (event) => {
            const payload = JSON.parse(event.data);

            if (payload.type === "message.created" && payload.data) {
              addMessage(payload.data);
              onDialogUpdated();
            }

            if (payload.type === "messages.snapshot" && Array.isArray(payload.data)) {
              setMessages((prev) => (prev.length > 0 ? prev : [...payload.data].reverse()));
            }
          };

          sock.onerror = (event) => {
            console.error("WebSocket Error:", event);
            setError("Ошибка соединения с WebSocket");
          };

          sock.onclose = (event) => {
            console.log("WebSocket connection closed:", event);
            setError("Соединение с WebSocket закрыто");
            if (!isCancelled) {
              setTimeout(connectSocket, 5000);
            }
          };
        };

        connectSocket();
      } catch (err: any) {
        if (isCancelled) return;

        if (err?.response?.status === 403) {
          setError("Нет доступа к этому диалогу");
        } else if (err?.response?.status === 404) {
          setError("Диалог не найден");
        } else {
          setError("Не удалось загрузить чат");
        }
      } finally {
        if (!isCancelled) setIsLoading(false);
      }
    }

    void loadDialogAndMessages();

    return () => {
      isCancelled = true;
      if (socketRef.current) {
        socketRef.current.close();
        socketRef.current = null;
      }
    };
  }, [dialogId, token, onDialogUpdated]);

  // Автоскролл вниз при новых сообщениях
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const canDelete = useMemo(
    () => dialog?.created_by === currentUserId,
    [dialog, currentUserId]
  );

  const handleSend = async () => {
    const content = text.trim();
    if (!content) return;

    try {
      const msg = await sendMessage(dialogId, content);
      // оптимистично покажем сообщение (хелпер предотвратит дубли при быстром broadcast).
      addMessage(msg);
      setText("");
      onDialogUpdated(); // обновить список диалогов
    } catch {
      setError("Не удалось отправить сообщение");
    }
  };

  const handleDeleteDialog = async () => {
    if (!canDelete) return;

    if (!window.confirm("Удалить этот чат? Действие необратимо.")) return;

    try {
      await deleteDialog(dialogId);
      onDialogDeleted();
    } catch (err: any) {
      if (err?.response?.status === 403) {
        setError("Удалить диалог может только создатель");
      } else {
        setError("Не удалось удалить диалог");
      }
    }
  };

  if (isLoading) {
    return <div className="chatPlaceholder">Загрузка чата...</div>;
  }

  if (error && !dialog) {
    return <div className="chatPlaceholder errorState">{error}</div>;
  }

  return (
    <div className="chatContainer">
      <header className="chatHeader">
        <div>
          <h2 className="chatTitle">
            {dialog ? getDialogTitle(dialog, currentUserId) : "Чат"}
          </h2>
          <div className="chatSubline">
            ID: {dialog?.id} · owner: {dialog?.created_by?.slice(0, 8) || "—"}
          </div>
        </div>

        {canDelete && (
          <button className="dangerButton" onClick={handleDeleteDialog}>
            Удалить чат
          </button>
        )}
      </header>

      <div className="chatMetaBar">
        <div>
          Участники:{" "}
          {(dialog?.participants ?? []).map((u) => u.username).join(", ") || "—"}
        </div>
        <button
          className="ghostButton small"
          onClick={() => navigator.clipboard.writeText(dialog?.id || "")}
        >
          Копировать ID
        </button>
      </div>

      {error && <div className="formError inline">{error}</div>}

      <div className="messagesViewport">
        {messages.length === 0 ? (
          <div className="placeholder">Сообщений пока нет</div>
        ) : (
          messages.map((msg) => {
            const isOwn = msg.sender_id === currentUserId;
            return (
              <div key={msg.id} className={`messageRow ${isOwn ? "mine" : "other"}`}>
                <div className={`messageBubble ${isOwn ? "mine" : "other"}`}>
                  <div className="messageText">{msg.content}</div>
                  <div className="messageMeta">
                    <span>{isOwn ? "Вы" : msg.sender_id.slice(0, 8)}</span>
                    <span>{formatTime(msg.created_at)}</span>
                  </div>
                </div>
              </div>
            );
          })
        )}
        <div ref={messagesEndRef} />
      </div>

      <div className="composer">
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && !e.shiftKey) {
              e.preventDefault();
              void handleSend();
            }
          }}
          placeholder="Введите сообщение..."
          rows={2}
        />
        <button className="primaryButton" onClick={handleSend}>
          Отправить
        </button>
      </div>
    </div>
  );
}