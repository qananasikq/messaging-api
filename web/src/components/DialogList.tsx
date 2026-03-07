import { useEffect, useMemo, useState } from "react";
import { createDialog, Dialog, getDialogs } from "../api";

function getDialogTitle(dialog: Dialog, currentUserId: string): string {
  if (dialog.name) return dialog.name;

  const otherParticipants = dialog.participants?.filter((p) => p.id !== currentUserId) ?? [];

  if (otherParticipants.length > 0) {
    return otherParticipants.map((p) => p.username).join(", ");
  }

  return dialog.type === "group" ? "Групповой чат" : "Личный чат";
}

type DialogListProps = {
  currentUserId: string;
  username: string;
  selectedDialogId: string | null;
  dialogsVersion: number;
  onOpen: (dialogId: string) => void;
  onLogout: () => void;
  onDialogsChanged: () => void;
};

export default function DialogList({
  currentUserId,
  username,
  selectedDialogId,
  dialogsVersion,
  onOpen,
  onLogout,
  onDialogsChanged,
}: DialogListProps) {
  const [dialogs, setDialogs] = useState<Dialog[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [createName, setCreateName] = useState("");
  const [participantIdsInput, setParticipantIdsInput] = useState("");
  const [openByIdInput, setOpenByIdInput] = useState("");
  const [error, setError] = useState("");
  const [isCreating, setIsCreating] = useState(false);

  useEffect(() => {
    void loadDialogs();
  }, [dialogsVersion]);

  async function loadDialogs() {
    setIsLoading(true);
    setError("");

    try {
      const data = await getDialogs();
      setDialogs(Array.isArray(data) ? data : []);
    } catch {
      setDialogs([]);
    } finally {
      setIsLoading(false);
    }
  }

  // Пока просто копируем массив — позже можно будет добавить сортировку по последнему сообщению
  const sortedDialogs = useMemo(() => [...dialogs], [dialogs]);

  async function handleCreateDialog() {
    setError("");

    const participantIds = participantIdsInput
      .split(",")
      .map((id) => id.trim())
      .filter(Boolean);

    if (participantIds.length === 0) {
      setError("Укажите хотя бы одного участника");
      return;
    }

    setIsCreating(true);

    try {
      const newDialog = await createDialog({
        name: createName.trim() || undefined,
        participant_ids: participantIds,
      });

      setCreateName("");
      setParticipantIdsInput("");
      onDialogsChanged();
      onOpen(newDialog.id);
    } catch (err: any) {
      if (err?.response?.status === 400) {
        setError("Проверьте формат UUID участников или название");
      } else {
        setError("Не удалось создать чат");
      }
    } finally {
      setIsCreating(false);
    }
  }

  function handleOpenById() {
    const trimmed = openByIdInput.trim();
    if (!trimmed) return;
    onOpen(trimmed);
    setOpenByIdInput(""); // опционально — очищаем поле после открытия
  }

  return (
    <div className="sidebarInner">
      <div className="sidebarTopCard">
        <div>
          <div className="profileOverline">Вы вошли как</div>
          <div className="profileName">{username}</div>
          <div className="profileMeta">ID: {currentUserId.slice(0, 8)}...</div>
        </div>
        <button className="ghostButton" onClick={onLogout}>
          Выйти
        </button>
      </div>

      <div className="sidebarBlock">
        <h3 className="blockTitle">Открыть по ID</h3>
        <input
          value={openByIdInput}
          onChange={(e) => setOpenByIdInput(e.target.value)}
          placeholder="Вставьте UUID диалога"
          onKeyDown={(e) => e.key === "Enter" && handleOpenById()}
        />
        <button className="primaryButton" onClick={handleOpenById}>
          Открыть
        </button>
      </div>

      <div className="sidebarBlock">
        <h3 className="blockTitle">Новый чат</h3>
        <input
          value={createName}
          onChange={(e) => setCreateName(e.target.value)}
          placeholder="Название (для группового чата)"
        />
        <textarea
          value={participantIdsInput}
          onChange={(e) => setParticipantIdsInput(e.target.value)}
          placeholder="UUID участников через запятую"
          rows={3}
        />
        <p className="blockHint">
          Ваш ID добавляется автоматически. Для личного чата достаточно одного чужого ID.
        </p>

        {error && <div className="formError compact">{error}</div>}

        <button
          className="primaryButton"
          onClick={handleCreateDialog}
          disabled={isCreating}
        >
          {isCreating ? "Создаётся..." : "Создать"}
        </button>
      </div>

      <div className="sidebarBlock fill">
        <div className="blockTitleRow">
          <h3 className="blockTitle">Чаты</h3>
          <button className="ghostButton small" onClick={() => void loadDialogs()}>
            ↻ Обновить
          </button>
        </div>

        {isLoading ? (
          <div className="placeholder">Загрузка...</div>
        ) : sortedDialogs.length === 0 ? (
          <div className="placeholder">Нет активных чатов</div>
        ) : (
          <div className="dialogList">
            {sortedDialogs.map((dialog) => {
              const isActive = selectedDialogId === dialog.id;
              const unread = dialog.unread_count ?? 0;

              return (
                <button
                  key={dialog.id}
                  className={`dialogCard ${isActive ? "active" : ""}`}
                  onClick={() => onOpen(dialog.id)}
                >
                  <div className="dialogCardHeader">
                    <div className="dialogCardTitle">
                      {getDialogTitle(dialog, currentUserId)}
                    </div>
                    {unread > 0 && (
                      <span className="unreadBadge">{unread}</span>
                    )}
                  </div>
                  <div className="dialogCardMeta">
                    {dialog.type === "group" ? "Группа" : "Личный"} ·{" "}
                    {dialog.id.slice(0, 8)}...
                  </div>
                  <div className="dialogCardPreview">
                    {dialog.last_message?.content || "Нет сообщений"}
                  </div>
                </button>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}