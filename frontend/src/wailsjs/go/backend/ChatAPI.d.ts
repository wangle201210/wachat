export interface Message {
  id: string;
  role: string;
  content: string;
  timestamp: string;
}

export interface Conversation {
  id: string;
  title: string;
  messages: Message[];
  createdAt: string;
  updatedAt: string;
}

export function CreateConversation(title: string): Promise<Conversation>;
export function GetConversation(id: string): Promise<Conversation>;
export function ListConversations(): Promise<Conversation[]>;
export function DeleteConversation(id: string): Promise<void>;
export function SendMessage(conversationID: string, content: string): Promise<Message>;
export function SendMessageStream(conversationID: string, content: string): Promise<void>;
export function UpdateConversationTitle(id: string, title: string): Promise<void>;
export function ClearMessages(id: string): Promise<void>;
