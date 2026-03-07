import axios from "axios"

export type AuthResponse = {
  id: string
  username: string
  token: string
}

export type UserMini = {
  id: string
  username: string
}

export type Message = {
  id: string
  dialog_id: string
  sender_id: string
  content: string
  created_at: string
}

export type MessageMini = {
  id: string
  sender_id: string
  content: string
  created_at: string
}

export type Dialog = {
  id: string
  type: "direct" | "group"
  name?: string | null
  created_by: string
  created_at: string
  participants?: UserMini[]
  last_message?: MessageMini | null
  unread_count?: number
}

export const api = axios.create({
  baseURL: "http://localhost:8080"
})

export function setToken(token: string | null) {
  if (token) {
    api.defaults.headers.common["Authorization"] = `Bearer ${token}`
    return
  }
  delete api.defaults.headers.common["Authorization"]
}

export async function register(username: string, password: string) {
  const res = await api.post<AuthResponse>("/users", { username, password })
  return res.data
}

export async function login(username: string, password: string) {
  const res = await api.post<AuthResponse>("/auth/login", { username, password })
  return res.data
}

export async function getDialogs() {
  const res = await api.get<{ dialogs: Dialog[] }>("/dialogs")
  return res.data.dialogs || []
}

export async function getDialog(dialogId: string) {
  const res = await api.get<Dialog>(`/dialogs/${dialogId}`)
  return res.data
}

export async function createDialog(payload: { name?: string; participant_ids: string[] }) {
  const res = await api.post<Dialog>("/dialogs", payload)
  return res.data
}

export async function deleteDialog(dialogId: string) {
  const res = await api.delete(`/dialogs/${dialogId}`)
  return res.data
}

export async function getMessages(dialogId: string) {
  const res = await api.get<{ messages: Message[] }>(`/dialogs/${dialogId}/messages`)
  return res.data.messages || []
}

export async function sendMessage(dialogId: string, content: string) {
  const res = await api.post<Message>("/messages", { dialog_id: dialogId, content })
  return res.data
}
