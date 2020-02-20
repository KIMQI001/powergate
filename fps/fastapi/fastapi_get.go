package fastapi

import (
	"bytes"
	"context"
	"fmt"
	"io"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-car"
	ipfsfiles "github.com/ipfs/go-ipfs-files"

	dstest "github.com/ipfs/go-merkledag/test"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/path"

	"github.com/ipfs/go-cid"
	ftypes "github.com/textileio/fil-tools/fps/types"
)

func (i *Instance) Get(ctx context.Context, c cid.Cid) (io.Reader, error) {
	ar := i.auditer.Start(ctx, i.ID().String())
	ar.Close()
	r, err := i.get(ctx, ar, c)
	if err != nil {
		ar.Errored(err)
		return nil, err
	}
	ar.Success()
	return r, nil
}

func (i *Instance) get(ctx context.Context, oa ftypes.OpAuditer, c cid.Cid) (io.Reader, error) {
	info, _, err := i.getCidInfo(c)
	checkErr(err)

	fmt.Printf("seems that saved cid has %d shards\n", len(info.Cold.Filecoin.Proposals))
	readers := make([]io.Reader, CantShards)
	for k, sh := range info.Cold.Filecoin.Proposals {
		if k >= CantShards {
			break
		}
		fmt.Println("retrieving shard ", sh.ShardCid)
		rc, err := i.dm.Retrieve(ctx, i.WalletAddr(), sh.ShardCid)
		checkErr(err)
		readers[k] = rc
	}
	mr := io.LimitReader(io.MultiReader(readers...), int64(info.Cold.Filecoin.CarSize))
	bserv := dstest.Bserv()
	_ = bserv
	_, err = car.LoadCar(&fakeStore{i.ipfs}, mr)
	checkErr(err)

	n, err := i.ipfs.Unixfs().Get(ctx, path.IpfsPath(c))
	if err != nil {
		return nil, err
	}
	file := ipfsfiles.ToFile(n)
	if file == nil {
		return nil, fmt.Errorf("node is a directory")
	}

	return file, nil
}

type fakeStore struct {
	ipfs iface.CoreAPI
}

func (fk *fakeStore) Put(b blocks.Block) error {
	fmt.Printf("fakestore: called put of block %v\n", b.Cid())
	_, err := fk.ipfs.Block().Put(context.Background(), bytes.NewReader(b.RawData()))
	checkErr(err)
	return nil
}
