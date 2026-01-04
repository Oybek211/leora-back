# Insights Backend Prompt

## Mission
Generate actionable insights for finance, planner, habit, and focus domains; expose CRUD for AI-assisted insights while ensuring every insight is tied to user context and relevance.

## Entities & Data Relationships
- `insights` hold `kind`, `level`, `scope`, `category`, `title`, `body`, `priority`, `payload`, and relational hints (`related` JSON). Each insight references a `user`.
- `insight_questions` and `insight_question_answers` support interactive Q&A flows; answers may create follow-up insights or adjust `insight` status.

## API Surface & Responsibilities
- GET `/insights`, `/insights/:id`; PATCH `/insights/:id/view`, `/dismiss`, `/complete` to change status flags.
- POST `/insights/generate` triggers AI generation (likely background job) with `scope`/`categories`; responses should include `payload` for templates.
- GET `/insights/history`, `/insights/questions`, POST `/insights/questions/:id/answer`, POST `/insights/ask` support conversational experiences.
- Status transitions should update `viewedAt`, `completedAt`, and `dismissedAt` while respecting soft deletion semantics.

## Real-Time & WebSocket Notes
- Push `insight:new` events through WebSocket whenever a new insight is generated (AI or rule-driven) so widgets and notifications surface them immediately.
- Status updates can also emit `entity:updated` for the `insights` entity.

## Operational Considerations
- Rate limit AI generation to 10 req/min and track usage via AI Usage module for quota enforcement.
- All insighs share `syncStatus`; when new insights arrive via WebSocket, clients may skip `sync/pull` for those records.
