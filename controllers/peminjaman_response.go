package controllers

import (
	"time"

	"backend/models"
)

// Bentuk response meniru struktur data contoh yang sudah dipakai frontend,
// supaya penyambungannya tidak perlu mengubah tampilan.

type peminjamanPeralatanResp struct {
	ID            uint    `json:"id"`
	NamaPeralatan string  `json:"nama_peralatan"`
	NomorAset     string  `json:"nomor_aset"`
	NomorSeri     string  `json:"nomor_seri"`
	Merek         string  `json:"merek"`
	Tipe          string  `json:"tipe"`
	LabPemilik    string  `json:"lab_pemilik"`
	TglJatuhTempo *string `json:"tgl_jatuh_tempo"`
}

type peminjamanPeminjamResp struct {
	UserID   uint64 `json:"user_id"`
	Nama     string `json:"nama"`
	NIP      string `json:"nip"`
	LabAsal  string `json:"lab_asal"`
	Operator string `json:"operator"`
}

// persetujuanResp adalah satu tahap persetujuan. Status berisi pending,
// approved, rejected, atau skipped. TTD selalu null karena persetujuan belum
// memakai tanda tangan.
type persetujuanResp struct {
	UserID  uint64     `json:"user_id"`
	Nama    string     `json:"nama"`
	NIP     string     `json:"nip"`
	Status  string     `json:"status"`
	Tanggal *time.Time `json:"tanggal"`
	Catatan string     `json:"catatan"`
	TTD     *string    `json:"ttd"`
}

type alatAlternatifResp struct {
	ID            uint   `json:"id"`
	NamaPeralatan string `json:"nama_peralatan"`
	NomorAset     string `json:"nomor_aset"`
}

type pengelolaResp struct {
	persetujuanResp
	AlatAlternatif *alatAlternatifResp `json:"alat_alternatif"`
}

// lampiranAResp adalah satu butir Lampiran A. Frontend membaca isinya berdasarkan
// urutan array, dengan field status dan keterangan.
type lampiranAResp struct {
	NomorButir int    `json:"nomor_butir"`
	Butir      string `json:"butir"`
	Status     string `json:"status"`
	Keterangan string `json:"keterangan"`
}

// serahTerimaResp adalah lembar serah terima keluar. Nilainya null sampai
// Pengelola menyimpan checklist. Tanggal diterima dan tanda tangan peminjam
// terisi setelah peminjam mengonfirmasi terima. Gambar tanda tangan hanya
// dikirim pada response detail agar daftar tetap ringan.
type serahTerimaResp struct {
	Tanggal          time.Time       `json:"tanggal"`
	PICPetugas       string          `json:"pic_petugas"`
	PeminjamPenerima string          `json:"peminjam_penerima"`
	TanggalDiterima  *time.Time      `json:"tanggal_diterima"`
	AdaTS            bool            `json:"ada_ts"`
	LampiranA        []lampiranAResp `json:"lampiran_a"`
	TTDPengelola     *string         `json:"ttd_pengelola"`
	TTDPeminjam      *string         `json:"ttd_peminjam"`
	Catatan          string          `json:"catatan"`
	F006Nomor        *string         `json:"f006_nomor"`
}

type peminjamanResp struct {
	ID               uint64    `json:"id"`
	Kode             string    `json:"kode"`
	TanggalPengajuan time.Time `json:"tanggal_pengajuan"`
	Status           string    `json:"status"`

	Peralatan peminjamanPeralatanResp `json:"peralatan"`
	Peminjam  peminjamanPeminjamResp  `json:"peminjam"`

	TujuanPenggunaan      string `json:"tujuan_penggunaan"`
	NomorSPK              string `json:"nomor_spk"`
	LokasiPenggunaan      string `json:"lokasi_penggunaan"`
	IsEksternal           bool   `json:"is_eksternal"`
	RencanaTanggalKeluar  string `json:"rencana_tanggal_keluar"`
	RencanaTanggalKembali string `json:"rencana_tanggal_kembali"`
	KebutuhanKelengkapan  string `json:"kebutuhan_kelengkapan"`
	KebutuhanAksesori     string `json:"kebutuhan_aksesori"`
	Catatan               string `json:"catatan"`

	ApprovalManagerPeminjam persetujuanResp `json:"approval_manager_peminjam"`
	ApprovalPengelola       pengelolaResp   `json:"approval_pengelola"`
	ApprovalManagerLab      persetujuanResp `json:"approval_manager_lab"`

	SerahTerimaKeluar   *serahTerimaResp `json:"serah_terima_keluar"`
	TanggalKeluarAktual *time.Time       `json:"tanggal_keluar_aktual"`
	AlasanPembatalan    string           `json:"alasan_pembatalan"`
}

func buildPersetujuan(
	user *models.User,
	userID uint64,
	status string,
	at *time.Time,
	catatan string,
) persetujuanResp {

	resp := persetujuanResp{
		UserID:  userID,
		Status:  status,
		Tanggal: at,
		Catatan: catatan,
	}

	if user != nil {
		resp.Nama = user.Name
		resp.NIP = user.NIP
	}

	return resp
}

// buildResponse menyusun response satu peminjaman untuk daftar dan hasil aksi.
// Gambar tanda tangan serah terima tidak disertakan.
func (c *PeminjamanController) buildResponse(p *models.Peminjaman) peminjamanResp {
	return c.susunResponse(p, false)
}

// buildResponseDetail sama seperti buildResponse, tetapi menyertakan gambar
// tanda tangan serah terima. Dipakai untuk response satu peminjaman.
func (c *PeminjamanController) buildResponseDetail(p *models.Peminjaman) peminjamanResp {
	return c.susunResponse(p, true)
}

// susunResponse menyusun response satu peminjaman. Relasi harus sudah dimuat,
// misalnya lewat repository FindByID atau FindAll.
func (c *PeminjamanController) susunResponse(p *models.Peminjaman, denganTTD bool) peminjamanResp {
	resp := peminjamanResp{
		ID:                    p.ID,
		Kode:                  p.Kode,
		TanggalPengajuan:      p.CreatedAt,
		Status:                p.Status,
		TujuanPenggunaan:      p.TujuanPenggunaan,
		NomorSPK:              p.NomorSPK,
		LokasiPenggunaan:      p.LokasiPenggunaan,
		IsEksternal:           p.IsEksternal,
		RencanaTanggalKeluar:  formatTanggalID(p.RencanaTanggalKeluar),
		RencanaTanggalKembali: formatTanggalID(p.RencanaTanggalKembali),
		KebutuhanKelengkapan:  p.KebutuhanKelengkapan,
		KebutuhanAksesori:     p.KebutuhanAksesori,
		Catatan:               p.Catatan,
		TanggalKeluarAktual:   p.TglKeluarAktual,
		AlasanPembatalan:      p.AlasanPembatalan,
	}

	if p.Peralatan != nil {
		resp.Peralatan = peminjamanPeralatanResp{
			ID:            p.Peralatan.ID,
			NamaPeralatan: p.Peralatan.NamaPeralatan,
			NomorAset:     p.Peralatan.NomorAset,
			NomorSeri:     p.Peralatan.NomorSeri,
			Merek:         p.Peralatan.Merek,
			Tipe:          p.Peralatan.TipeModel,
		}

		if jatuhTempo := c.jatuhTempoPeralatan(p.Peralatan); jatuhTempo != nil {
			text := formatTanggalID(*jatuhTempo)
			resp.Peralatan.TglJatuhTempo = &text
		}
	}

	if p.LabPemilik != nil {
		resp.Peralatan.LabPemilik = p.LabPemilik.NamaLabs
	}

	resp.Peminjam = peminjamanPeminjamResp{
		UserID:   p.PeminjamID,
		Operator: p.NamaOperator,
	}

	if p.Peminjam != nil {
		resp.Peminjam.Nama = p.Peminjam.Name
		resp.Peminjam.NIP = p.Peminjam.NIP
	}

	if p.LabPeminjam != nil {
		resp.Peminjam.LabAsal = p.LabPeminjam.NamaLabs
	}

	resp.ApprovalManagerPeminjam = buildPersetujuan(
		p.ManajerPeminjam, p.ManajerPeminjamID,
		p.ManajerPeminjamStatus, p.ManajerPeminjamAt, p.ManajerPeminjamCatatan,
	)

	resp.ApprovalPengelola = pengelolaResp{
		persetujuanResp: buildPersetujuan(
			p.Pengelola, p.PengelolaID,
			p.PengelolaStatus, p.PengelolaAt, p.PengelolaCatatan,
		),
	}

	if p.AlatAlternatif != nil {
		resp.ApprovalPengelola.AlatAlternatif = &alatAlternatifResp{
			ID:            p.AlatAlternatif.ID,
			NamaPeralatan: p.AlatAlternatif.NamaPeralatan,
			NomorAset:     p.AlatAlternatif.NomorAset,
		}
	}

	resp.ApprovalManagerLab = buildPersetujuan(
		p.ManagerLab, p.ManagerLabID,
		p.ManagerLabStatus, p.ManagerLabAt, p.ManagerLabCatatan,
	)

	if p.SerahTerima != nil {
		st := p.SerahTerima

		serah := &serahTerimaResp{
			Tanggal:         st.DiperiksaAt,
			TanggalDiterima: st.DiterimaAt,
			AdaTS:           st.AdaTS,
			Catatan:         st.Catatan,
			LampiranA:       make([]lampiranAResp, 0, len(st.Items)),
		}

		if p.Pengelola != nil {
			serah.PICPetugas = p.Pengelola.Name
		}

		if p.Peminjam != nil {
			serah.PeminjamPenerima = p.Peminjam.Name
		}

		for _, item := range st.Items {
			serah.LampiranA = append(serah.LampiranA, lampiranAResp{
				NomorButir: item.NomorButir,
				Butir:      item.Butir,
				Status:     item.Hasil,
				Keterangan: item.Keterangan,
			})
		}

		if denganTTD {
			if st.PengelolaTTD != "" {
				ttd := st.PengelolaTTD
				serah.TTDPengelola = &ttd
			}

			if st.PeminjamTTD != "" {
				ttd := st.PeminjamTTD
				serah.TTDPeminjam = &ttd
			}
		}

		resp.SerahTerimaKeluar = serah
	}

	return resp
}
