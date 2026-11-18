# Offline map preview

Developer switch for checking client data without a server. Nothing is sent:
the network client is never connected.

```sh
GORO_PREVIEW_MAP=rag_fes,200,200        # map[,x,y[,job]]  (default 150,150, Novice)
GORO_PREVIEW_ACTORS=1002,1115,4054,4218 # optional job/monster/NPC IDs placed in rows above the player
goro --data-dir /path/to/data --no-audio --windowed
```

The game skips login and enters the map. It still loads the map, its models and textures, the sprites, the lighting and the water. It does not run server packets, so there are no NPCs, warps or other players beyond what you place with `GORO_PREVIEW_ACTORS`. The switch is ignored with `--headless`.

`scripts/data/preview-shots.sh` takes screenshots on a machine without a GPU
by running the client under Xvfb with the Mesa lavapipe Vulkan driver. Set
`VK_ICD_FILENAMES` to the lavapipe ICD if more than one driver is installed.
