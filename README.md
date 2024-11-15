# ceph-quickrbytes

qickly polls all subdirs under a given parent dir for CephFS recursive byte usage 

apx 10x faster than looping calls of `getfattr`
apx 100x faster than `du`

usage:
`ceph-quickrbytes /mnt/cephfs/parent_dir`

example:

    user@host~# /root/ceph-quickrbytes /mnt/ceph/quota_test/home/ | tail -n5
    /mnt/ceph/quota_test/home/user995       3761242112
    /mnt/ceph/quota_test/home/user996       1862270976
    /mnt/ceph/quota_test/home/user997       7627341824
    /mnt/ceph/quota_test/home/user998       6083837952
    /mnt/ceph/quota_test/home/user999       2994733056


example with output units:

    user@host~# /root/ceph-quickrbytes /mnt/ceph/quota_test/home/ | tail -n5
    /mnt/ceph/quota_test/home/user995       3.50 GB
    /mnt/ceph/quota_test/home/user996       1.73 GB
    /mnt/ceph/quota_test/home/user997       7.10 GB
    /mnt/ceph/quota_test/home/user998       5.67 GB
    /mnt/ceph/quota_test/home/user999       2.79 GB


