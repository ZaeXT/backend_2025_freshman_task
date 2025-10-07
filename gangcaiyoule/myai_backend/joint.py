import os
import math
import argparse
from PIL import Image
import colorsys

SUPPORTED_FORMATS = ('.png', '.jpg', '.jpeg')

def average_rgb(img: Image.Image, resize_to=(50, 50)):
    img = img.convert('RGB').resize(resize_to)
    pixels = list(img.getdata())
    r = sum(p[0] for p in pixels) / len(pixels)
    g = sum(p[1] for p in pixels) / len(pixels)
    b = sum(p[2] for p in pixels) / len(pixels)
    return (r, g, b)

def rgb_to_hsv(rgb):
    return colorsys.rgb_to_hsv(*(v / 255 for v in rgb))

def merge_by_hue(output_dir, save_path="merged_by_hue.png"):
    image_files = [os.path.join(output_dir, f)
                   for f in os.listdir(output_dir)
                   if f.lower().endswith(SUPPORTED_FORMATS)]
    images_with_hue = []
    for f in image_files:
        img = Image.open(f).convert('RGB')
        avg_rgb = average_rgb(img)
        hue = rgb_to_hsv(avg_rgb)[0]
        images_with_hue.append((img, hue))
    images_sorted = sorted(images_with_hue, key=lambda x: x[1])
    img_width, img_height = images_sorted[0][0].size
    max_per_row = int(math.sqrt(len(images_sorted)))
    num_rows = math.ceil(len(images_sorted) / max_per_row)
    total_width = img_width * max_per_row
    total_height = img_height * num_rows
    merged_image = Image.new('RGB', (total_width, total_height), color=(255, 255, 255))
    for idx, (img, _) in enumerate(images_sorted):
        row = idx // max_per_row
        col = idx % max_per_row
        x = col * img_width
        y = row * img_height
        merged_image.paste(img, (x, y))
    merged_image.save(save_path)
    print(f"已按色相排序并拼接，共 {len(images_sorted)} 张图，保存为 {save_path}")

def make_mosaic(tiles_dir, target_img_path, save_path="mosaic_result.jpg", grid_size=80):
    tile_files = [os.path.join(tiles_dir, f)
                  for f in os.listdir(tiles_dir)
                  if f.lower().endswith(SUPPORTED_FORMATS)]
    def avg_rgb_tile(img):
        return average_rgb(img, resize_to=(30, 30))
    tiles = []
    for f in tile_files:
        img = Image.open(f).convert('RGB')
        avg = avg_rgb_tile(img)
        tiles.append((img, avg))
    print(f"共加载 {len(tiles)} 张头像作为素材")
    target_img = Image.open(target_img_path).convert("RGB")
    target_img = target_img.resize((64*20, 64*20))
    cell_w = target_img.width // grid_size
    cell_h = target_img.height // grid_size
    mosaic = Image.new('RGB', (target_img.width, target_img.height))
    for row in range(grid_size):
        for col in range(grid_size):
            x = col * cell_w
            y = row * cell_h
            cell = target_img.crop((x, y, x+cell_w, y+cell_h))
            cell_avg = average_rgb(cell, resize_to=(cell_w, cell_h))
            best_img = min(tiles, key=lambda t: sum((a-b)**2 for a, b in zip(t[1], cell_avg)))[0]
            mosaic.paste(best_img.resize((cell_w, cell_h)), (x, y))
    mosaic.save(save_path)
    print(f"马赛克拼图已完成，保存为 {save_path}")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="动漫头像图片工具集")
    parser.add_argument("--output_dir", type=str, default="./output", help="动漫头像图片文件夹")
    parser.add_argument("--mode", type=str, choices=["merge", "mosaic"], required=True, help="功能选择: merge=色相拼接, mosaic=马赛克拼图")
    parser.add_argument("--target", type=str, help="用于马赛克拼图的目标图片路径 (mode=mosaic 时必填)")
    parser.add_argument("--grid_size", type=int, default=80, help="马赛克拼图的网格数量 (默认80, mode=mosaic时有效)")
    parser.add_argument("--save", type=str, default=None, help="保存文件名")
    args = parser.parse_args()

    if args.mode == "merge":
        save_path = args.save or "merged_by_hue.png"
        merge_by_hue(args.output_dir, save_path=save_path)
    elif args.mode == "mosaic":
        if not args.target:
            print("请指定 --target 目标图片进行马赛克拼图")
            exit(1)
        save_path = args.save or "mosaic_result.jpg"
        make_mosaic(args.output_dir, args.target, save_path=save_path, grid_size=args.grid_size)