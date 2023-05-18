package gredis

import (
	"bufio"
	"context"
)

type cmdable func(ctx context.Context, cmd Cmder) error

func NewCmdfunc(cn *Conn) cmdable {
	return func(ctx context.Context, cmd Cmder) error {
		if err := cn.WithWriter(context.Background(), func(wr *bufio.Writer) error {
			return writeCmd(wr, cmd)
		}); err != nil {
			return err
		}

		if err := cn.WithReader(context.Background(), cmd.readReply); err != nil {
			return err
		}

		return nil
	}
}

func (c cmdable) StrSet(ctx context.Context, key string, value string) *StatusCmd {
	args := make([]interface{}, 3, 5)
	args[0] = "set"
	args[1] = key
	args[2] = value

	cmd := NewStatusCmd(ctx, args...)
	_ = c(ctx, cmd)
	return cmd
}

func (c cmdable) StrGet(ctx context.Context, key string) *StringCmd {
	args := make([]interface{}, 2, 5)
	args[0] = "get"
	args[1] = key

	cmd := NewStringCmd(ctx, args...)
	_ = c(ctx, cmd)
	return cmd
}
