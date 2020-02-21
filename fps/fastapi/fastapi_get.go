package fastapi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"time"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-car"
	ipfsfiles "github.com/ipfs/go-ipfs-files"
	"github.com/logrusorgru/aurora"

	dstest "github.com/ipfs/go-merkledag/test"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/klauspost/reedsolomon"

	"github.com/ipfs/go-cid"
	ftypes "github.com/textileio/fil-tools/fps/types"
)

var SuccessProb = 0.8

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

	cant := 0
	dataShardsUsed := 0
	parityShardsUsed := 0
	readers := make([][]byte, CantShards+CantParity)
	for k, sh := range info.Cold.Filecoin.Proposals {
		if cant >= CantShards {
			break
		}
		if rand.Float64() > SuccessProb {
			continue
		}
		if k >= CantShards {
			parityShardsUsed++
		} else {
			dataShardsUsed++
		}
		rc, err := i.dm.Retrieve(ctx, i.WalletAddr(), sh.ShardCid)
		checkErr(err)
		readers[k], err = ioutil.ReadAll(rc)
		if err != nil {
			return nil, err
		}
		cant++
	}

	if cant != CantShards {
		return nil, fmt.Errorf("Can't reconstruct data, %d (<%d) shards available for reconstruction", cant, CantShards)
	}

	fmt.Printf(aurora.Sprintf(aurora.Magenta("Reconstructing data using %d data shards and %d parity shards...\n"), dataShardsUsed, parityShardsUsed))
	start := time.Now()
	if parityShardsUsed > 0 {
		enc, err := reedsolomon.New(CantShards, CantParity)
		checkErr(err)
		err = enc.ReconstructData(readers)
		if err != nil {
			return nil, fmt.Errorf("reconstructing data failed: %s", err)
		}
	}

	fmt.Printf(aurora.Sprintf(aurora.Magenta("Reconstructing data took %.2f ms\n"), float64(time.Since(start).Microseconds())/float64(1000)))

	readers2 := make([]io.Reader, CantShards)
	for k, r := range readers {
		if k >= CantShards {
			break
		}
		readers2[k] = bytes.NewReader(r)
	}

	mr := io.LimitReader(io.MultiReader(readers2...), int64(info.Cold.Filecoin.CarSize))
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
	_, err := fk.ipfs.Block().Put(context.Background(), bytes.NewReader(b.RawData()))
	checkErr(err)
	return nil
}

func Message(format string, args ...interface{}) {
	fmt.Println(aurora.Sprintf(aurora.BrightBlack("> "+format), args...))
}
