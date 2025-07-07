package bot

import "github.com/celestix/gotgproto/ext"

func (b *Bot) handleStart(ctx *ext.Context, update *ext.Update) error {
	_, err := ctx.Reply(update, ext.ReplyTextString("hello"), &ext.ReplyOpts{})
	if err != nil {
		return err
	}

	return nil
}
