package bot

import "github.com/celestix/gotgproto/ext"

func (b *Bot) handleStart(ctx *ext.Context, update *ext.Update) error {
	_, err := ctx.Reply(update, ext.ReplyTextString("Привет, я Лотти! Бот для модерации плейлистов."+"\n\n"+
		"Добавь меня в группу и я подгружу данные из неё, если хочешь приватный плейлист только для тебя - зайди в миниапп."+"\n\n"+
		"Для управления плейлистами, к которым у тебя есть доступ - зайди в миниапп)"), &ext.ReplyOpts{})
	if err != nil {
		b.logger.Error(err.Error())
		return nil
	}

	return nil
}
